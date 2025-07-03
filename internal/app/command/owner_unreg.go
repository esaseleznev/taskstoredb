package command

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/serialx/hashring"
)

type OwnerUnRegDbAdapter interface {
	OwnerUnReg(owner string) (events []contract.Event, err error)
	Apply(events []contract.Event) (err error)
}

type OwnerUnRegClusterAdapter interface {
	OwnerUnReg(url string, owner string) (err error)
}

type OwnerUnRegHandler struct {
	db      OwnerUnRegDbAdapter
	cluster OwnerUnRegClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
	raft    *raft.Raft
}

func NewOwnerUnRegHandler(
	db OwnerUnRegDbAdapter,
	cluster OwnerUnRegClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
	raft *raft.Raft,
) OwnerUnRegHandler {
	if db == nil {
		panic("nil OwnerUnRegDbAdapter")
	}
	if cluster == nil {
		panic("nil OwnerUnRegClusterAdapter")
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

	return OwnerUnRegHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		nodes:   nodes,
		raft:    raft,
	}
}

func (h OwnerUnRegHandler) Handle(owner string, internal bool) (err error) {
	if owner == "" {
		return errors.New("owner is empty")
	}

	if internal {
		events, err := h.db.OwnerUnReg(owner)
		if err != nil {
			return err
		}
		return raftApply(h.raft, h.db, events)
	}

	for _, node := range h.nodes {
		if node == h.curUrl {
			events, err := h.db.OwnerUnReg(owner)
			if err != nil {
				return err
			}
			err = raftApply(h.raft, h.db, events)
		} else {
			err = h.cluster.OwnerUnReg(node, owner)
		}
		if err != nil {
			return err
		}
	}
	return err
}
