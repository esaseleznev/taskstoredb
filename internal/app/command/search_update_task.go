package command

import (
	"errors"
	"maps"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/serialx/hashring"
)

type SearchUpdateTaskDbAdapter interface {
	SearchTask(
		condition *contract.Condition,
		kind *string,
		size *uint,
	) (tasks []contract.Task, err error)

	Update(
		id string,
		status contract.Status,
		param map[string]string,
		error *string,
		offset *string,
	) (events []contract.Event, err error)

	Delete(id string) (events []contract.Event, err error)

	Apply(events []contract.Event) (err error)
}

type SearchUpdateTaskClusterAdapter interface {
	SearchUpdateTask(
		url string,
		up contract.TaskUpdate,
		condition *contract.Condition,
		kind *string,
		size *uint,
	) (err error)

	Add(
		url string,
		group string,
		kind string,
		owner *string,
		param map[string]string,
	) (id string, err error)
}

type SearchUpdateTaskHandler struct {
	db      SearchUpdateTaskDbAdapter
	cluster SearchUpdateTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	nodes   []string
	raft    *raft.Raft
}

func NewSearchUpdateTaskHandler(
	db SearchUpdateTaskDbAdapter,
	cluster SearchUpdateTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	nodes []string,
	raft *raft.Raft,
) (h SearchUpdateTaskHandler, err error) {
	if db == nil {
		return h, errors.New("nil SearchUpdateTaskDbAdapter")
	}
	if cluster == nil {
		return h, errors.New("nil SearchUpdateTaskClusterAdapter")
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

	return SearchUpdateTaskHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		nodes:   nodes,
		raft:    raft,
	}, nil
}

func (h SearchUpdateTaskHandler) Handle(
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
			err = h.cluster.SearchUpdateTask(node, up, condition, kind, size)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h SearchUpdateTaskHandler) internal(
	up contract.TaskUpdate,
	condition *contract.Condition,
	kind *string,
	size *uint,
) (err error) {
	portion, err := h.db.SearchTask(condition, kind, size)
	if err != nil {
		return err
	}
	for _, task := range portion {
		isNew := false
		if up.Status != nil {
			task.Status = *up.Status
		}
		if up.Param != nil {
			maps.Copy(task.Param, up.Param)
		}
		if up.Error != nil {
			task.Error = up.Error
		}
		if up.Kind != nil {
			task.Kind = *up.Kind
			isNew = true
		}
		if up.Group != nil {
			task.Group = *up.Group
			isNew = true
		}
		if up.Owner != nil {
			task.Owner = up.Owner
			isNew = true
		}

		if !isNew {
			events, err := h.db.Update(task.Id, task.Status, task.Param, task.Error, nil)
			if err != nil {
				return err
			}
			err = raftApply(h.raft, h.db, events)
			if err != nil {
				return err
			}
		} else {
			// order is important to not lose the task
			// if the outcome is bad there may be a duplicate
			_, err = h.cluster.Add(h.curUrl, task.Group, task.Kind, task.Owner, task.Param)
			if err != nil {
				return err
			}
			events, err := h.db.Delete(task.Id)
			if err != nil {
				return err
			}
			err = raftApply(h.raft, h.db, events)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
