package query

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type SearchErrorTaskDbAdapter interface {
	SearchErrorTask(condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error)
}

type SearchErrorTaskClusterAdapter interface {
	SearchErrorTask(url string, condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error)
}

type SearchErrorTaskHandler struct {
	db      SearchErrorTaskDbAdapter
	cluster SearchErrorTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
}

func NewSearchErrorTaskHandler(
	db SearchErrorTaskDbAdapter,
	cluster SearchErrorTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) SearchErrorTaskHandler {
	if db == nil {
		panic("nil SearchErrorTaskDbAdapter")
	}
	if cluster == nil {
		panic("nil SearchErrorTaskClusterAdapter")
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

	return SearchErrorTaskHandler{db: db, cluster: cluster, ring: ring, curUrl: url, nodes: nodes}
}

func (h SearchErrorTaskHandler) Handle(condition *contract.Condition, kind *string, size *uint, internal bool) (tasks []contract.Task, err error) {
	if condition != nil && len(condition.Operations) == 0 && len(condition.Operations) == 0 {
		return tasks, errors.New("condition is empty")
	}

	if internal {
		return h.db.SearchErrorTask(condition, kind, size)
	}

	var portion []contract.Task

	for _, node := range h.nodes {
		if node == h.curUrl {
			portion, err = h.db.SearchErrorTask(condition, kind, size)
		} else {
			portion, err = h.cluster.SearchErrorTask(node, condition, kind, size)
		}
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, portion...)
	}
	return tasks, err
}
