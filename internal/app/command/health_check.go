package command

import (
	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type HealthCheckDbAdapter interface {
	HealthCheck() (events []contract.Event, err error)
	Apply(events []contract.Event) (err error)
}

type HealthCheckHandler struct {
	db   HealthCheckDbAdapter
	ring *hashring.HashRing
}

func NewHealthCheckHandler(
	db HealthCheckDbAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) HealthCheckHandler {
	if db == nil {
		panic("nil HealthCheckDbAdapter")
	}
	if ring == nil {
		panic("nil ring")
	}
	if url == "" {
		panic("url is empty")
	}
	if len(nodes) == 0 {
		panic("nodes is empty")
	}

	return HealthCheckHandler{db: db, ring: ring}
}

func (h HealthCheckHandler) Handle() (err error) {
	events, err := h.db.HealthCheck()
	if err != nil {
		return err
	}
	return h.db.Apply(events)
}
