package command

type AddTask struct {
	Group string
	Param map[string]string
	Kind  string
}

type AddTaskRepository interface {
	Add(group string, kind string, param map[string]string) (id string, err error)
}

type AddTaskHandler struct {
	repo AddTaskRepository
}

func NewAddTaskHandler(repo AddTaskRepository) AddTaskHandler {
	if repo == nil {
		panic("nil addTaskRepository")
	}

	return AddTaskHandler{repo: repo}
}

func (h AddTaskHandler) Handle(cmd AddTask) (id string, err error) {
	return h.repo.Add(cmd.Group, cmd.Kind, cmd.Param)
}
