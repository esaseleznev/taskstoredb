package query

import (
	"errors"
	"fmt"

	"github.com/serialx/hashring"
)

type GetFirstInGroupDbAdapter interface {
	GetFirstInGroup(group string) (id string, err error)
}

type GetFirstInGroupClusterAdapter interface {
	GetFirstInGroup(url string, group string) (id string, err error)
}

type GetFirstInGroupHandler struct {
	db      GetFirstInGroupDbAdapter
	cluster GetFirstInGroupClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
}

func NewGetFirstInGroupHandler(
	db GetFirstInGroupDbAdapter,
	cluster GetFirstInGroupClusterAdapter,
	ring *hashring.HashRing,
	url string,
) GetFirstInGroupHandler {
	if db == nil {
		panic("nil GetFirstInGroupDbAdapter")
	}
	if cluster == nil {
		panic("nil GetFirstInGroupClusterAdapter")
	}
	if ring == nil {
		panic("nil ring")
	}
	if url == "" {
		panic("url is empty")
	}

	return GetFirstInGroupHandler{db: db, cluster: cluster, ring: ring, curUrl: url}
}

func (h GetFirstInGroupHandler) Handle(group string) (id string, err error) {
	if group == "" {
		return id, errors.New("group is empty")
	}

	node, exists := h.ring.GetNode(group)
	if !exists {
		return id, fmt.Errorf("not found node by group: %v", node)
	}

	if node == h.curUrl {
		return h.db.GetFirstInGroup(group)
	} else {
		return h.cluster.GetFirstInGroup(node, group)
	}
}
