package main

import (
	"log"
	"os"

	"github.com/esaseleznev/taskstoredb/internal/adapters"
	"github.com/esaseleznev/taskstoredb/internal/app"
	"github.com/esaseleznev/taskstoredb/internal/app/command"
	"github.com/esaseleznev/taskstoredb/internal/app/query"
	"github.com/esaseleznev/taskstoredb/internal/config"
	hport "github.com/esaseleznev/taskstoredb/internal/ports/http"
	"github.com/serialx/hashring"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	config := config.NewConfig()
	application := newApplication(config, logger)
	httpServer := hport.NewHttpServer(config.Http.Port, application, logger)
	httpServer.Start()
}

func newApplication( /*ctx context.Context,*/ config config.Config, logger *log.Logger) app.Application {
	level, err := leveldb.OpenFile(config.Db.Path, nil)
	if err != nil {
		logger.Fatalf("Could not open leveldb %+v\n", err)
	}
	db := adapters.NewLevelAdapter(level)
	cluster := adapters.HttpClusterAdapter{}

	servers := config.Cluster.Servers
	ring := hashring.New(servers)

	return app.Application{
		Commands: app.Commands{
			AddTask:    command.NewAddTaskHandler(db, cluster, ring, config.Cluster.Current),
			UpdateTask: command.NewUpdateTaskHendler(db, cluster, ring, config.Cluster.Current),
			OwnerReg:   command.NewOwnerRegHandler(db, cluster, ring, config.Cluster.Current, servers),
			SetOffset:  command.NewSetOffsetHandler(db, cluster, ring, config.Cluster.Current, servers),
		},
		Queries: app.Queries{
			GetFirstInGroup: query.NewGetFirstInGroupHandler(db, cluster, ring, config.Cluster.Current),
		},
	}
}
