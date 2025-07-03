package command

import (
	"errors"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/serialx/hashring"
)

type SearchDeleteErrorTaskDbAdapter interface {
	SearchErrorTask(
		condition *contract.Condition,
		kind *string,
		size *uint,
	) (tasks []contract.Task, err error)
	DeleteError(id string) (events []contract.Event, err error)
	Apply(events []contract.Event) (err error)
}

type SearchDeleteErrorTaskClusterAdapter interface {
	SearchDeleteErrorTask(
		url string,
		condition *contract.Condition,
		kind *string,
		size *uint,
	) (err error)
}

type SearchDeleteErrorTaskHandler struct {
	db      SearchDeleteErrorTaskDbAdapter
	cluster SearchDeleteErrorTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
	raft    *raft.Raft
}

func NewSearchDeleteErrorTaskHandler(
	db SearchDeleteErrorTaskDbAdapter,
	cluster SearchDeleteErrorTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
	raft *raft.Raft,
) (h SearchDeleteErrorTaskHandler, err error) {
	if db == nil {
		return h, errors.New("nil SearchDeleteErrorTaskDbAdapter")
	}
	if cluster == nil {
		return h, errors.New("nil SearchDeleteErrorTaskClusterAdapter")
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

	return SearchDeleteErrorTaskHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		nodes:   nodes,
		raft:    raft,
	}, nil
}

func (h SearchDeleteErrorTaskHandler) Handle(
	condition *contract.Condition,
	kind *string,
	size *uint,
	internal bool,
) (err error) {
	if condition != nil && len(condition.Operations) == 0 && len(condition.Conditions) == 0 {
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
			err = h.cluster.SearchDeleteErrorTask(node, condition, kind, size)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h SearchDeleteErrorTaskHandler) internal(
	condition *contract.Condition,
	kind *string,
	size *uint,
) (err error) {
	portion, err := h.db.SearchErrorTask(condition, kind, size)
	if err != nil {
		return err
	}
	for _, task := range portion {
		events, err := h.db.DeleteError(task.Id)
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
