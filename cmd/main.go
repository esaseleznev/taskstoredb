package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"time"

	cluster "github.com/esaseleznev/taskstoredb/internal/adapters/cluster/http"
	store "github.com/esaseleznev/taskstoredb/internal/adapters/store/leveldb"
	"github.com/esaseleznev/taskstoredb/internal/app"
	"github.com/esaseleznev/taskstoredb/internal/app/command"
	"github.com/esaseleznev/taskstoredb/internal/app/query"
	"github.com/esaseleznev/taskstoredb/internal/config"
	hport "github.com/esaseleznev/taskstoredb/internal/ports/http"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
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
	httpServer := hport.NewHttpServer(config.Cluster.CurrentPort, application, logger)
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

	raft, err := newRaft(&config, (*store.Fsm)(db))
	if err != nil {
		logger.Fatalf("failed to create raft: %v", err)
	}

	return app.Application{
		Commands: app.Commands{
			AddTask:               command.NewAddTaskHandler(db, cluster, ring, config.Cluster.Current, raft),
			UpdateTask:            command.NewUpdateTaskHendler(db, cluster, ring, config.Cluster.Current, raft),
			OwnerReg:              command.NewOwnerRegHandler(db, cluster, ring, config.Cluster.Current, servers, raft),
			OwnerUnReg:            command.NewOwnerUnRegHandler(db, cluster, ring, config.Cluster.Current, servers, raft),
			SearchDeleteTask:      command.NewSearchDeleteTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft),
			SearchDeleteErrorTask: command.NewSearchDeleteErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft),
			SearchUpdateTask:      command.NewSearchUpdateTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft),
			SearchUpdateErrorTask: command.NewSearchUpdateErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft),
			HealthCheck:           command.NewHealthCheckHandler(db, raft),
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

func newRaft(config *config.Config, fsm *store.Fsm) (*raft.Raft, error) {
	os.MkdirAll(config.Raft.Path, os.ModePerm)

	store, err := raftboltdb.NewBoltStore(path.Join(config.Raft.Path, "bolt"))
	if err != nil {
		return nil, fmt.Errorf("Could not create bolt store: %s", err)
	}

	snapshots, err := raft.NewFileSnapshotStore(path.Join(config.Raft.Path, "snapshot"), 2, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("Could not create snapshot store: %s", err)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Raft.Current.Address)
	if err != nil {
		return nil, fmt.Errorf("Could not resolve address: %s", err)
	}

	transport, err := raft.NewTCPTransport(
		config.Raft.Current.Address,
		tcpAddr,
		10,
		time.Second*10,
		os.Stderr,
	)
	if err != nil {
		return nil, fmt.Errorf("Could not create tcp transport: %s", err)
	}

	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(config.Raft.Current.Id)

	r, err := raft.NewRaft(raftCfg, fsm, store, store, snapshots, transport)
	if err != nil {
		return nil, fmt.Errorf("Could not create raft instance: %s", err)
	}

	servers := []raft.Server{
		{
			ID:      raft.ServerID(config.Raft.Current.Id),
			Address: transport.LocalAddr(),
		},
	}
	for _, server := range config.Raft.Servers {
		servers = append(servers, raft.Server{
			ID:      raft.ServerID(server.Id),
			Address: raft.ServerAddress(server.Address),
		})
	}

	r.BootstrapCluster(raft.Configuration{
		Servers: servers,
	},
	)

	return r, nil
}
