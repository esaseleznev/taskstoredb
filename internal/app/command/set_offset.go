package command

import (
	"github.com/serialx/hashring"
)

type SetOffsetDbAdapter interface {
	SetOffset(owner string, kind string, startId string) (err error)
}

type SetOffsetClusterAdapter interface {
	SetOffset(url string, owner string, kind string, startId string) (err error)
}

type SetOffsetHandler struct {
	db      SetOffsetDbAdapter
	cluster SetOffsetClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
}

func NewSetOffsetHandler(
	db SetOffsetDbAdapter,
	cluster SetOffsetClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) SetOffsetHandler {
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

	return SetOffsetHandler{db: db, cluster: cluster, ring: ring, curUrl: url, nodes: nodes}
}

func (h SetOffsetHandler) Handle(owner string, kind string, startId string, internal bool) (err error) {
	if internal {
		return h.db.SetOffset(owner, kind, startId)
	}

	for _, node := range h.nodes {
		if node == h.curUrl {
			err = h.db.SetOffset(owner, kind, startId)
		} else {
			err = h.cluster.SetOffset(node, owner, kind, startId)
		}
		if err != nil {
			return err
		}
	}
	return err
}
