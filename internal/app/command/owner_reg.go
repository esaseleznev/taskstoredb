package command

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/serialx/hashring"
)

type OwnerRegDbAdapter interface {
	OwnerReg(owner string, kinds []string) (events []contract.Event)
	Apply(events []contract.Event) (err error)
}

type OwnerRegClusterAdapter interface {
	OwnerReg(url string, owner string, kinds []string) (err error)
}

type OwnerRegHandler struct {
	db      OwnerRegDbAdapter
	cluster OwnerRegClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
	raft    *raft.Raft
}

func NewOwnerRegHandler(
	db OwnerRegDbAdapter,
	cluster OwnerRegClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
	raft *raft.Raft,
) (h OwnerRegHandler, err error) {
	if db == nil {
		return h, errors.New("nil ownerRegDbAdapter")
	}
	if cluster == nil {
		return h, errors.New("nil ownerRegClusterAdapter")
	}
	if ring == nil {
		return h, errors.New("nil ring")
	}
	if url == "" {
		return h, errors.New("url is empty")
	}
	if len(nodes) == 0 {
		return h, errors.New("nodes is empty")
	}

	return OwnerRegHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		nodes:   nodes,
		raft:    raft,
	}, nil
}

func (h OwnerRegHandler) Handle(
	owner string,
	kinds []string,
	internal bool,
) (err error) {
	if owner == "" {
		return errors.New("owner is empty")
	}
	if len(kinds) == 0 {
		return errors.New("kinds is empty")
	}

	if internal {
		events := h.db.OwnerReg(owner, kinds)
		return raftApply(h.raft, h.db, events)

	}

	for _, node := range h.nodes {
		if node == h.curUrl {
			events := h.db.OwnerReg(owner, kinds)
			err = raftApply(h.raft, h.db, events)
		} else {
			err = h.cluster.OwnerReg(node, owner, kinds)
		}
		if err != nil {
			return err
		}
	}
	return err
}
