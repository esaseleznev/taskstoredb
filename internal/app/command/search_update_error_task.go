package command

import (
	"errors"

	"maps"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type SearchUpdateErrorTaskDbAdapter interface {
	SearchErrorTask(condition *contract.Condition, kind *string, size *uint) (tasks []contract.Task, err error)
	UpdateError(
		id string,
		status contract.Status,
		param map[string]string,
	) (err error)
}

type SearchUpdateErrorTaskClusterAdapter interface {
	SearchUpdateErrorTask(url string, up contract.TaskUpdate, condition *contract.Condition, kind *string, size *uint) (err error)
}

type SearchUpdateErrorTaskHandler struct {
	db      SearchUpdateErrorTaskDbAdapter
	cluster SearchUpdateErrorTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
}

func NewSearchUpdateErrorTaskHandler(
	db SearchUpdateErrorTaskDbAdapter,
	cluster SearchUpdateErrorTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
) SearchUpdateErrorTaskHandler {
	if db == nil {
		panic("nil SearchUpdateErrorTaskDbAdapter")
	}
	if cluster == nil {
		panic("nil SearchUpdateErrorTaskClusterAdapter")
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

	return SearchUpdateErrorTaskHandler{db: db, cluster: cluster, ring: ring, curUrl: url, nodes: nodes}
}

func (h SearchUpdateErrorTaskHandler) Handle(
	up contract.TaskUpdate,
	condition *contract.Condition,
	kind *string,
	size *uint,
	internal bool,
) (err error) {
	if condition != nil && len(condition.Operations) == 0 && len(condition.Operations) == 0 {
		return errors.New("condition is empty")
	}

	if internal {
		return h.internal(up, condition, kind, size)
	}

	for _, node := range h.nodes {
		if node == h.curUrl {
			err = h.internal(up, condition, kind, size)
			if err != nil {
				return err
			}
		} else {
			err = h.cluster.SearchUpdateErrorTask(node, up, condition, kind, size)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h SearchUpdateErrorTaskHandler) internal(
	up contract.TaskUpdate,
	condition *contract.Condition,
	kind *string,
	size *uint,
) (err error) {
	portion, err := h.db.SearchErrorTask(condition, kind, size)
	if err != nil {
		return err
	}
	for _, task := range portion {
		if up.Status != nil {
			task.Status = *up.Status
		}
		if up.Param != nil {
			maps.Copy(task.Param, up.Param)
		}

		err = h.db.UpdateError(task.Id, task.Status, task.Param)
		if err != nil {
			return err
		}
	}
	return nil
}
