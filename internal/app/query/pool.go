package query

import (
	"errors"
	"sort"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

const (
	size uint = 1000
)

type PoolDbAdapter interface {
	Pool(
		owner string,
		kind string,
		size uint,
	) (tasks []contract.Task, err error)
}

type PoolClusterAdapter interface {
	Pool(
		url string,
		owner string,
		kind string,
	) (tasks []contract.Task, err error)
}

type PoolHandler struct {
	db      PoolDbAdapter
	cluster PoolClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
}

func NewPoolHandler(
	db PoolDbAdapter,
	cluster PoolClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) PoolHandler {
	if db == nil {
		panic("nil poolDbAdapter")
	}
	if cluster == nil {
		panic("nil poolClusterAdapter")
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

	return PoolHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		nodes:   nodes,
	}
}

func (h PoolHandler) Handle(
	owner string,
	kind string,
	internal bool,
) (tasks []contract.Task, err error) {
	if owner == "" {
		return tasks, errors.New("owner is empty")
	}
	if kind == "" {
		return tasks, errors.New("kind is empty")
	}

	if internal {
		return h.db.Pool(owner, kind, size)
	}

	var portion []contract.Task

	for _, node := range h.nodes {
		if node == h.curUrl {
			portion, err = h.db.Pool(owner, kind, size)
		} else {
			portion, err = h.cluster.Pool(node, owner, kind)
		}
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, portion...)
	}

	sort.SliceStable(tasks, func(i int, j int) bool {
		return tasks[i].Id < tasks[j].Id
	})

	return tasks, err
}
