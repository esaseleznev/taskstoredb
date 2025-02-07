package ports

import (
	"net/http"

	"github.com/esaseleznev/taskstoredb/internal/app/command"
)

type AddTask struct {
	Group string            `json:"g"`
	Param map[string]string `json:"p"`
	Kind  string            `json:"k"`
}

func (h HttpServer) Add() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := decode[AddTask](r)
		if err != nil {
			_ = encode(w, int(http.StatusBadRequest), NewErrorResult(err))
			return
		}

		id, err := h.app.Commands.AddTask.Handle(command.AddTask{Group: t.Group, Param: t.Param, Kind: t.Kind})
		if err != nil {
			_ = encode(w, int(http.StatusBadRequest), NewErrorResult(err))
			return
		}

		encode(w, int(http.StatusOK), NewIdResult(id))
	}
}
