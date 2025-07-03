package command

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
)

type HealthCheckDbAdapter interface {
	HealthCheck() (events []contract.Event, err error)
	Apply(events []contract.Event) (err error)
}

type HealthCheckHandler struct {
	db   HealthCheckDbAdapter
	raft *raft.Raft
}

func NewHealthCheckHandler(
	db HealthCheckDbAdapter,
	raft *raft.Raft,
) (h HealthCheckHandler, err error) {
	if db == nil {
		return h, errors.New("nil HealthCheckDbAdapter")
	}

	return HealthCheckHandler{db: db, raft: raft}, nil
}

func (h HealthCheckHandler) Handle() (err error) {
	events, err := h.db.HealthCheck()
	if err != nil {
		return err
	}
	return raftApply(h.raft, h.db, events)
}
