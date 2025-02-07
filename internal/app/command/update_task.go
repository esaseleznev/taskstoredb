package command

import (
	entity "github.com/esaseleznev/taskstoredb/internal/domain"
)

type UpdateTask struct {
	Id     string
	Status entity.Status
	Param  map[string]string
	Error  string
}

type UpdateTaskRepository interface {
	//get(id string) (task entity.Task, err error)
	Update(id string, status entity.Status, param map[string]string, error *string) (err error)
}

type UpdateTaskHendler struct {
	repo UpdateTaskRepository
}

func NewUpdateTaskHendler(repo UpdateTaskRepository) UpdateTaskHendler {
	if repo == nil {
		panic("nil updateTaskRepository")
	}

	return UpdateTaskHendler{repo: repo}
}

func (h UpdateTaskHendler) Handle(cmd UpdateTask) (err error) {
	return h.repo.Update(cmd.Id, cmd.Status, cmd.Param, nil)
}
