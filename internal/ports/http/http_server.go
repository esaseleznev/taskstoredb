package ports

import (
	"context"
	"encoding/json"
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
)

type key int

const (
	requestIDKey key = 0
)

type HttpServer struct {
	port    string
	app     app.Application
	logger  *log.Logger
	healthy int32
}

type ErrorResult struct {
	Error string `json:"error"`
}

type IdResult struct {
	Id string `json:"id"`
}

func NewErrorResult(err error) ErrorResult {
	return ErrorResult{
		Error: err.Error(),
	}
}

func NewIdResult(id string) IdResult {
	return IdResult{
		Id: id,
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
	http.HandleFunc("POST /task", h.Add())

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

func (h HttpServer) HeathCheck() http.HandlerFunc {
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
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
