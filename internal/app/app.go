package app

import (
	"github.com/esaseleznev/taskstoredb/internal/app/command"
	"github.com/esaseleznev/taskstoredb/internal/app/query"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	AddTask          command.AddTaskHandler
	UpdateTask       command.UpdateTaskHendler
	OwnerReg         command.OwnerRegHandler
	SetOffset        command.SetOffsetHandler
	SearchDeleteTask command.SearchDeleteTaskHandler
}

type Queries struct {
	GetFirstInGroup query.GetFirstInGroupHandler
	Pool            query.PoolHandler
	Get             query.GetHandler
	SearchTask      query.SearchTaskHandler
}
