package main

import (
	"log"
	"net/http"
	"os"
	"time"

	cluster "github.com/esaseleznev/taskstoredb/internal/adapters/cluster/http"
	store "github.com/esaseleznev/taskstoredb/internal/adapters/store/leveldb"
	"github.com/esaseleznev/taskstoredb/internal/app"
	"github.com/esaseleznev/taskstoredb/internal/app/command"
	"github.com/esaseleznev/taskstoredb/internal/app/query"
	"github.com/esaseleznev/taskstoredb/internal/config"
	hport "github.com/esaseleznev/taskstoredb/internal/ports/http"
	"github.com/justinrixx/retryhttp"
	"github.com/serialx/hashring"
	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	config, err := config.NewConfig(logger)
	if err != nil {
		logger.Fatalf("Could not create config %+v\n", err)
	}
	application := newApplication(config, logger)
	httpServer := hport.NewHttpServer(config.Http.Port, application, logger)
	httpServer.Start()
}

func newApplication( /*ctx context.Context,*/ config config.Config, logger *log.Logger) app.Application {
	level, err := leveldb.OpenFile(config.Db.Path, nil)
	if err != nil {
		logger.Fatalf("Could not open leveldb %+v\n", err)
	}

	db, err := store.NewLevelAdapter(level)
	if err != nil {
		logger.Fatalf("Could not create level adapter %+v\n", err)
	}

	httpClient := &http.Client{
		Transport: retryhttp.New(
			// optional retry configurations
			retryhttp.WithShouldRetryFn(func(attempt retryhttp.Attempt) bool {
				return attempt.Res != nil && attempt.Res.StatusCode == http.StatusServiceUnavailable
			}),
			retryhttp.WithDelayFn(func(attempt retryhttp.Attempt) time.Duration {
				return time.Duration(attempt.Count*3) * time.Second
			}),
			retryhttp.WithMaxRetries(3),
		),
		// other HTTP client options
	}
	cluster := cluster.NewHttpClusterAdapter(httpClient)

	servers := config.Cluster.Servers
	ring := hashring.New(servers)

	return app.Application{
		Commands: app.Commands{
			AddTask:               command.NewAddTaskHandler(db, cluster, ring, config.Cluster.Current),
			UpdateTask:            command.NewUpdateTaskHendler(db, cluster, ring, config.Cluster.Current),
			OwnerReg:              command.NewOwnerRegHandler(db, cluster, ring, config.Cluster.Current, servers),
			OwnerUnReg:            command.NewOwnerUnRegHandler(db, cluster, ring, config.Cluster.Current, servers),
			SearchDeleteTask:      command.NewSearchDeleteTaskHandler(db, cluster, ring, config.Cluster.Current, servers),
			SearchDeleteErrorTask: command.NewSearchDeleteErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers),
			SearchUpdateTask:      command.NewSearchUpdateTaskHandler(db, cluster, ring, config.Cluster.Current, servers),
			SearchUpdateErrorTask: command.NewSearchUpdateErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers),
		},
		Queries: app.Queries{
			GetFirstInGroup: query.NewGetFirstInGroupHandler(db, cluster, ring, config.Cluster.Current),
			Pool:            query.NewPoolHandler(db, cluster, ring, config.Cluster.Current, servers),
			Get:             query.NewGetHandler(db, cluster, ring, config.Cluster.Current),
			SearchTask:      query.NewSearchTaskHandler(db, cluster, ring, config.Cluster.Current, servers),
			SearchError:     query.NewSearchErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers),
		},
	}
}
