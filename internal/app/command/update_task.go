package command

import (
	"errors"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/hashicorp/raft"
	"github.com/serialx/hashring"
)

type UpdateTaskDbAdapter interface {
	Update(
		id string,
		status contract.Status,
		param map[string]string,
		error *string,
		offset *string,
	) (events []contract.Event, err error)
	Apply(events []contract.Event) (err error)
}

type UpdateTaskClusterAdapter interface {
	Update(
		url string,
		group string,
		id string,
		status contract.Status,
		param map[string]string,
		error *string,
	) (err error)
}

type UpdateTaskHandler struct {
	db      UpdateTaskDbAdapter
	cluster UpdateTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
	raft    *raft.Raft
}

func NewUpdateTaskHandler(
	db UpdateTaskDbAdapter,
	cluster UpdateTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
	raft *raft.Raft,
) (h UpdateTaskHandler, err error) {
	if db == nil {
		return h, errors.New("nil updateTaskAdapter")
	}
	if cluster == nil {
		return h, errors.New("nil updateTaskClusterAdapter")
	}
	if ring == nil {
		return h, errors.New("nil ring")
	}
	if url == "" {
		return h, errors.New("url is empty")
	}

	return UpdateTaskHandler{
		db:      db,
		cluster: cluster,
		ring:    ring,
		curUrl:  url,
		raft:    raft,
	}, nil
}

func (h UpdateTaskHandler) Handle(
	group string,
	id string,
	status contract.Status,
	param map[string]string,
	error *string,
) (err error) {
	if group == "" {
		return errors.New("group is empty")
	}
	if id == "" {
		return errors.New("id is empty")
	}
	if status == 0 {
		return errors.New("status is empty")
	}

	node, exists := h.ring.GetNode(group)
	if !exists {
		return fmt.Errorf("not found node by group: %v", node)
	}

	if node == h.curUrl {
		var offset *string
		if status == contract.COMPLETED || status == contract.FAILED {
			offset = &id
		}
		events, err := h.db.Update(id, status, param, error, offset)
		if err != nil {
			return err
		}
		return raftApply(h.raft, h.db, events)
	} else {
		return h.cluster.Update(node, group, id, status, param, error)
	}
}
