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
	AddTask               command.AddTaskHandler
	UpdateTask            command.UpdateTaskHandler
	OwnerReg              command.OwnerRegHandler
	OwnerUnReg            command.OwnerUnRegHandler
	SearchDeleteTask      command.SearchDeleteTaskHandler
	SearchDeleteErrorTask command.SearchDeleteErrorTaskHandler
	SearchUpdateTask      command.SearchUpdateTaskHandler
	SearchUpdateErrorTask command.SearchUpdateErrorTaskHandler
	HealthCheck           command.HealthCheckHandler
}

type Queries struct {
	GetFirstInGroup query.GetFirstInGroupHandler
	Pool            query.PoolHandler
	Get             query.GetHandler
	SearchTask      query.SearchTaskHandler
	SearchError     query.SearchErrorTaskHandler
}
