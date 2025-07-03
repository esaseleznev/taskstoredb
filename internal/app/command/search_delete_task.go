package command

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/serialx/hashring"
)

type SearchDeleteTaskDbAdapter interface {
	SearchTask(
		condition *contract.Condition,
		kind *string,
		size *uint,
	) (tasks []contract.Task, err error)
	Delete(id string) (events []contract.Event, err error)
	Apply(events []contract.Event) (err error)
}

type SearchDeleteTaskClusterAdapter interface {
	SearchDeleteTask(
		url string,
		condition *contract.Condition,
		kind *string,
		size *uint,
	) (err error)
}

type SearchDeleteTaskHandler struct {
	db      SearchDeleteTaskDbAdapter
	cluster SearchDeleteTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
	raft    *raft.Raft
}

func NewSearchDeleteTaskHandler(
	db SearchDeleteTaskDbAdapter,
	cluster SearchDeleteTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
	raft *raft.Raft,
) SearchDeleteTaskHandler {
	if db == nil {
		panic("nil SearchDeleteTaskDbAdapter")
	}
	if cluster == nil {
		panic("nil SearchDeleteTaskClusterAdapter")
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

	return SearchDeleteTaskHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		nodes:   nodes,
		raft:    raft,
	}
}

func (h SearchDeleteTaskHandler) Handle(
	condition *contract.Condition,
	kind *string,
	size *uint,
	internal bool,
) (err error) {
	if condition != nil && len(condition.Operations) == 0 && len(condition.Operations) == 0 {
		return errors.New("condition is empty")
	}

	if internal {
		return h.internal(condition, kind, size)
	}

	for _, node := range h.nodes {
		if node == h.curUrl {
			err = h.internal(condition, kind, size)
			if err != nil {
				return err
			}
		} else {
			err = h.cluster.SearchDeleteTask(node, condition, kind, size)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h SearchDeleteTaskHandler) internal(condition *contract.Condition, kind *string, size *uint) (err error) {
	portion, err := h.db.SearchTask(condition, kind, size)
	if err != nil {
		return err
	}
	for _, task := range portion {
		events, err := h.db.Delete(task.Id)
		if err != nil {
			return err
		}
		err = raftApply(h.raft, h.db, events)
		if err != nil {
			return err
		}
	}
	return nil
}
