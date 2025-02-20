package ports

import (
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/domain"
)

func (h HttpServer) Add() handlerFunc {
	type request struct {
		Group string            `json:"g"`
		Param map[string]string `json:"p"`
		Kind  string            `json:"k"`
	}
	type response struct {
		Id string `json:"id"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		t, err := decode[request](r)
		if err != nil {
			return err
		}

		id, err := h.app.Commands.AddTask.Handle(t.Group, t.Kind, t.Param)
		if err != nil {
			return err
		}

		return encode(w, int(http.StatusOK), response{Id: id})
	}
}

func (h HttpServer) Update() handlerFunc {
	type request struct {
		Id     string            `json:"id"`
		Group  string            `json:"g"`
		Status int               `json:"s"`
		Param  map[string]string `json:"p"`
		Error  *string           `json:"e"`
	}
	return func(w http.ResponseWriter, r *http.Request) error {
		t, err := decode[request](r)
		if err != nil {
			return err
		}

		err = h.app.Commands.UpdateTask.Handle(t.Id, t.Group, domain.Status(t.Status), t.Param, t.Error)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}
