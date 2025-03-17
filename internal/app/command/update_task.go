package command

import (
	"errors"
	"fmt"

	"github.com/esaseleznev/taskstoredb/internal/contract"
	"github.com/serialx/hashring"
)

type UpdateTaskDbAdapter interface {
	Update(
		id string,
		status contract.Status,
		param map[string]string,
		error *string,
	) (err error)
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

type UpdateTaskHendler struct {
	db      UpdateTaskDbAdapter
	cluster UpdateTaskClusterAdapter
	ring    *hashring.HashRing
	curUrl  string
}

func NewUpdateTaskHendler(
	db UpdateTaskDbAdapter,
	cluster UpdateTaskClusterAdapter,
	ring *hashring.HashRing,
	url string,
) UpdateTaskHendler {
	if db == nil {
		panic("nil updateTaskAdapter")
	}
	if cluster == nil {
		panic("nil updateTaskClusterAdapter")
	}
	if ring == nil {
		panic("nil ring")
	}
	if url == "" {
		panic("url is empty")
	}

	return UpdateTaskHendler{db: db, cluster: cluster, ring: ring, curUrl: url}
}

func (h UpdateTaskHendler) Handle(
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
		return h.db.Update(id, status, param, error)
	} else {
		return h.cluster.Update(node, group, id, status, param, error)
	}
}
