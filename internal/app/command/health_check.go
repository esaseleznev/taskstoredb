package command

import (
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
) HealthCheckHandler {
	if db == nil {
		panic("nil HealthCheckDbAdapter")
	}

	return HealthCheckHandler{db: db, raft: raft}
}

func (h HealthCheckHandler) Handle() (err error) {
	events, err := h.db.HealthCheck()
	if err != nil {
		return err
	}
	return raftApply(h.raft, h.db, events)
}
