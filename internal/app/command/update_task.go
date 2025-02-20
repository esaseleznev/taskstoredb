package command

import (
	"fmt"

	entity "github.com/esaseleznev/taskstoredb/internal/domain"
	"github.com/serialx/hashring"
)

type UpdateTask struct {
	Group  string
	Id     string
	Kind   string
	Status entity.Status
	Param  map[string]string
	Error  *string
}

type UpdateTaskDbAdapter interface {
	Update(
		id string,
		status entity.Status,
		param map[string]string,
		error *string,
	) (err error)
}

type UpdateTaskClusterAdapter interface {
	Update(
		url string,
		group string,
		id string,
		status entity.Status,
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
		panic("nil updateTaskDbAdapter")
	}
	if cluster == nil {
		panic("nil updateTaskClusterAdapter")
	}
	if url == "" {
		panic("url is empty")
	}

	return UpdateTaskHendler{db: db, cluster: cluster, ring: ring, curUrl: url}
}

func (h UpdateTaskHendler) Handle(
	group string,
	id string,
	status entity.Status,
	param map[string]string,
	error *string,
) (err error) {
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
