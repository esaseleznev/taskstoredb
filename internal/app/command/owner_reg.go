package command

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
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
}

func NewOwnerRegHandler(
	db OwnerRegDbAdapter,
	cluster OwnerRegClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) OwnerRegHandler {
	if db == nil {
		panic("nil ownerRegDbAdapter")
	}
	if cluster == nil {
		panic("nil ownerRegClusterAdapter")
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

	return OwnerRegHandler{db: db, cluster: cluster, ring: ring, curUrl: url, nodes: nodes}
}

func (h OwnerRegHandler) Handle(owner string, kinds []string, internal bool) (err error) {
	if owner == "" {
		return errors.New("owner is empty")
	}
	if len(kinds) == 0 {
		return errors.New("kinds is empty")
	}

	if internal {
		events := h.db.OwnerReg(owner, kinds)
		return h.db.Apply(events)

	}

	for _, node := range h.nodes {
		if node == h.curUrl {
			events := h.db.OwnerReg(owner, kinds)
			err = h.db.Apply(events)
		} else {
			err = h.cluster.OwnerReg(node, owner, kinds)
		}
		if err != nil {
			return err
		}
	}
	return err
}
