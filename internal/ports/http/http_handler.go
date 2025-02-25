package ports

import (
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/contract"
)

func (h HttpServer) Add() handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		t, err := decode[contract.AddRequest](r)
		if err != nil {
			return err
		}

		id, err := h.app.Commands.AddTask.Handle(t.Group, t.Kind, t.Param)
		if err != nil {
			return err
		}

		return encode(w, int(http.StatusOK), contract.AddResponse{Id: id})
	}
}

func (h HttpServer) Update() handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		t, err := decode[contract.UpdateRequest](r)
		if err != nil {
			return err
		}

		err = h.app.Commands.UpdateTask.Handle(
			t.Id,
			t.Group,
			contract.Status(t.Status),
			t.Param,
			t.Error,
		)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		return nil
	}
}
