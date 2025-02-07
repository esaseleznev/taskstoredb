package app

import "github.com/esaseleznev/taskstoredb/internal/app/command"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	AddTask    command.AddTaskHandler
	UpdateTask command.UpdateTaskHendler
}

type Queries struct {
}
