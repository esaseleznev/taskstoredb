package command

import (
	"errors"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
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
	raft    *raft.Raft
}

func NewAddTaskHandler(
	db AddTaskDbAdapter,
	cluster AddTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	raft *raft.Raft,
) (h AddTaskHandler, err error) {
	if db == nil {
		return h, errors.New("nil AddTaskDbAdapter")
	}
	if cluster == nil {
		return h, errors.New("nil AddTaskClusterAdapter")
	}
	if ring == nil {
		return h, errors.New("nil ring")
	}
	if url == "" {
		return h, errors.New("url is empty")
	}

	return AddTaskHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		raft:    raft,
	}, nil
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
		err = raftApply(h.raft, h.db, events)
		if err != nil {
			return id, err
		}
		id = string(events[0].Key)
		return id, nil
	} else {
		return h.cluster.Add(node, group, kind, owner, param)
	}
}
