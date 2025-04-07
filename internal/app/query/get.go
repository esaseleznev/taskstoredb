package query

import (
	"errors"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type GetDbAdapter interface {
	Get(id string) (tasks *contract.Task, err error)
}

type GetClusterAdapter interface {
	Get(url string, group string, id string) (tasks *contract.Task, err error)
}

type GetHandler struct {
	db      GetDbAdapter
	cluster GetClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
}

func NewGetHandler(
	db GetDbAdapter,
	cluster GetClusterAdapter,
	ring *hashring.HashRing,
	url string,
) GetHandler {
	if db == nil {
		panic("nil GetDbAdapter")
	}
	if cluster == nil {
		panic("nil GetClusterAdapter")
	}
	if ring == nil {
		panic("nil ring")
	}
	if url == "" {
		panic("url is empty")
	}

	return GetHandler{db: db, cluster: cluster, ring: ring, curUrl: url}
}

func (h GetHandler) Handle(group string, id string) (task *contract.Task, err error) {
	if group == "" {
		return task, errors.New("group is empty")
	}
	if id == "" {
		return task, errors.New("id is empty")
	}

	node, exists := h.ring.GetNode(group)
	if !exists {
		return task, fmt.Errorf("not found node by group: %v", node)
	}

	if node == h.curUrl {
		return h.db.Get(id)
	} else {
		return h.cluster.Get(node, group, id)
	}
}
