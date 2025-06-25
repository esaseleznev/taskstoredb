package command

import (
	"errors"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type AddTaskDbAdapter interface {
	Add(
		group string,
		kind string,
		owner *string,
		param map[string]string,
	) (events []contract.Event, err error)

	Apply(events []contract.Event) (err error)
}

type AddTaskClusterAdapter interface {
	Add(
		url string,
		group string,
		kind string,
		owner *string,
		param map[string]string,
	) (id string, err error)
}

type AddTaskHandler struct {
	db      AddTaskDbAdapter
	cluster AddTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
}

func NewAddTaskHandler(
	db AddTaskDbAdapter,
	cluster AddTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
) AddTaskHandler {
	if db == nil {
		panic("nil addTaskDbAdapter")
	}
	if cluster == nil {
		panic("nil AddTaskClusterAdapter")
	}
	if ring == nil {
		panic("nil ring")
	}
	if url == "" {
		panic("url is empty")
	}

	return AddTaskHandler{db: db, cluster: cluster, ring: ring, curUrl: url}
}

func (h AddTaskHandler) Handle(
	group string,
	kind string,
	owner *string,
	param map[string]string,
) (id string, err error) {
	if group == "" {
		return id, errors.New("group is empty")
	}
	if kind == "" {
		return id, errors.New("kind is empty")
	}

	node, exists := h.ring.GetNode(group)
	if !exists {
		return id, fmt.Errorf("not found node by group: %v", node)
	}

	if node == h.curUrl {
		events, err := h.db.Add(group, kind, owner, param)
		if err != nil {
			return id, err
		}
		err = h.db.Apply(events)
		if err != nil {
			return id, err
		}
		id = string(events[0].Key)
		return id, nil
	} else {
		return h.cluster.Add(node, group, kind, owner, param)
	}
}
