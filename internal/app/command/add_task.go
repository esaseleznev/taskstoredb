package command

import (
	"errors"
	"fmt"

	"github.com/serialx/hashring"
)

type AddTaskDbAdapter interface {
	Add(
		group string,
		kind string,
		owner *string,
		param map[string]string,
	) (id string, err error)
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
		return h.db.Add(group, kind, owner, param)
	} else {
		return h.cluster.Add(node, group, kind, owner, param)
	}
}
