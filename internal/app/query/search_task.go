package query

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type SearchTaskDbAdapter interface {
	SearchTask(condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error)
}

type SearchTaskClusterAdapter interface {
	SearchTask(url string, condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error)
}

type SearchTaskHandler struct {
	db      SearchTaskDbAdapter
	cluster SearchTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
}

func NewSearchTaskHandler(
	db SearchTaskDbAdapter,
	cluster SearchTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) SearchTaskHandler {
	if db == nil {
		panic("nil SearchTaskDbAdapter")
	}
	if cluster == nil {
		panic("nil SearchTaskClusterAdapter")
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

	return SearchTaskHandler{db: db, cluster: cluster, ring: ring, curUrl: url, nodes: nodes}
}

func (h SearchTaskHandler) Handle(condition *contract.Condition, kind *string, size *uint, internal bool) (tasks []contract.Task, err error) {
	if condition != nil && len(condition.Operations) == 0 && len(condition.Operations) == 0 {
		return tasks, errors.New("condition is empty")
	}

	if internal {
		return h.db.SearchTask(condition, kind, size)
	}

	var portion []contract.Task

	for _, node := range h.nodes {
		if node == h.curUrl {
			portion, err = h.db.SearchTask(condition, kind, size)
		} else {
			portion, err = h.cluster.SearchTask(node, condition, kind, size)
		}
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, portion...)
	}
	return tasks, err
}
