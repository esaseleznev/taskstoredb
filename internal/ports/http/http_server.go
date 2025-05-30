package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/esaseleznev/taskstoredb/internal/app"
	"github.com/esaseleznev/taskstoredb/internal/contract"
)

type key int

const (
	requestIDKey key = 0
)

type handlerFunc func(a app.Application, w http.ResponseWriter, r *http.Request) error

type HttpServer struct {
	port    string
	app     app.Application
	logger  *log.Logger
	healthy int32
}

type HttpError struct {
	Msg    string
	Status int
}

func (e HttpError) Error() string {
	return e.Msg
}

func (h HttpServer) handle(f handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(h.app, w, r); err != nil {
			status := http.StatusInternalServerError
			var httpError HttpError
			if errors.As(err, &httpError) {
				status = httpError.Status
			}

			if err := encode(w, int(status), NewErrorResult(err)); err != nil {
				log.Printf("failed to encode error: %s\n", err)
			}
		}
	}
}

func NewErrorResult(err error) contract.ErrorResponse {
	return contract.ErrorResponse{
		Error: err.Error(),
	}
}

func NewHttpServer(port string, application app.Application, logger *log.Logger) HttpServer {
	return HttpServer{
		port:   port,
		app:    application,
		logger: logger,
	}
}

func (h *HttpServer) Start() {
	http.HandleFunc("GET /healthz", h.HeathCheck())
	http.HandleFunc("POST /task", h.handle(Add))
	http.HandleFunc("PATCH /task", h.handle(Update))
	http.HandleFunc("PUT /owner", h.handle(OwnerReg))
	http.HandleFunc("GET /task/{id}/group/{group}", h.handle(Get))
	http.HandleFunc("GET /task/group/{group}", h.handle(GetFirstInGroup))
	http.HandleFunc("GET /pool/{owner}/kind/{kind}", h.handle(Pool))
	http.HandleFunc("POST /task/search", h.handle(SearchTask))
	http.HandleFunc("POST /error/search", h.handle(SearchError))
	http.HandleFunc("POST /task/search/delete", h.handle(SearchDeleteTask))
	http.HandleFunc("POST /error/search/delete", h.handle(SearchDeleteErrorTask))
	http.HandleFunc("POST /task/search/update", h.handle(SearchUpdateTask))
	http.HandleFunc("POST /error/search/update", h.handle(SearchUpdateErrorTask))

	nextRequestID := func() string {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	server := &http.Server{
		Addr:         ":" + h.port,
		Handler:      h.Tracing(nextRequestID)(h.Logging(h.logger)(http.DefaultServeMux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// set server to healthy
	h.setHealthy(1)
	go func() {
		<-quit
		log.Println("Server is shutting down ....")
		h.setHealthy(0)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		defer cancel()
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			h.logger.Fatalf("Could not gracefully shutdown the server %+v\n", err)
		}
		close(done)
	}()

	h.logger.Printf("Server starting at port %v ...", h.port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		h.logger.Fatalf("Could not listen on :%v %+v\n", h.port, err)
	}

	<-done
	h.logger.Println("Server stopped")
}

func (h *HttpServer) setHealthy(val int32) {
	atomic.StoreInt32(&h.healthy, val)
}

func (h *HttpServer) HeathCheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadInt32(&h.healthy) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (h HttpServer) Tracing(nextReuestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextReuestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h HttpServer) Logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode request: %w", err)
	}
	return v, nil
}
