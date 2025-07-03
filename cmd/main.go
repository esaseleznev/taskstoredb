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
		logger.Printf("Could not create config %+v\n", err)
		return
	}
	application, err := newApplication(config)
	if err != nil {
		logger.Printf("Could not create application %+v\n", err)
		return
	}
	httpServer := hport.NewHttpServer(config.Cluster.CurrentPort, application, logger)
	err = httpServer.Start()
	if err != nil {
		logger.Printf("Http server fatal error %+v\n", err)
		return
	}
}

func newApplication( /*ctx context.Context,*/ config config.Config) (a app.Application, err error) {
	level, err := leveldb.OpenFile(config.Db.Path, nil)
	if err != nil {
		return a, fmt.Errorf("Could not open leveldb %+v\n", err)
	}

	db, err := store.NewLevelAdapter(level)
	if err != nil {
		return a, fmt.Errorf("Could not create level adapter %+v\n", err)
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
		return a, fmt.Errorf("failed to create raft: %v", err)
	}

	addTask, err := command.NewAddTaskHandler(db, cluster, ring, config.Cluster.Current, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create add task handler: %v", err)
	}

	updateTask, err := command.NewUpdateTaskHandler(db, cluster, ring, config.Cluster.Current, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create update task handler: %v", err)
	}

	ownerReg, err := command.NewOwnerRegHandler(db, cluster, ring, config.Cluster.Current, servers, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create owner registration handler: %v", err)
	}

	ownerUnReg, err := command.NewOwnerUnRegHandler(db, cluster, ring, config.Cluster.Current, servers, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create owner unregistration handler: %v", err)
	}

	searchDeleteTask, err := command.NewSearchDeleteTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create search delete task handler: %v", err)
	}

	searchDeleteErrorTask, err := command.NewSearchDeleteErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create search delete error task handler: %v", err)
	}

	searchUpdateTask, err := command.NewSearchUpdateTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create search update task handler: %v", err)
	}

	searchUpdateErrorTask, err := command.NewSearchUpdateErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create search update error task handler: %v", err)
	}

	healthCheck, err := command.NewHealthCheckHandler(db, raft)
	if err != nil {
		return a, fmt.Errorf("failed to create health check handler: %v", err)
	}

	getFirstInGroup, err := query.NewGetFirstInGroupHandler(db, cluster, ring, config.Cluster.Current)
	if err != nil {
		return a, fmt.Errorf("failed to create get first in group handler: %v", err)
	}

	pool, err := query.NewPoolHandler(db, cluster, ring, config.Cluster.Current, servers)
	if err != nil {
		return a, fmt.Errorf("failed to create pool handler: %v", err)
	}

	get, err := query.NewGetHandler(db, cluster, ring, config.Cluster.Current)
	if err != nil {
		return a, fmt.Errorf("failed to create get handler: %v", err)
	}

	searchTask, err := query.NewSearchTaskHandler(db, cluster, ring, config.Cluster.Current, servers)
	if err != nil {
		return a, fmt.Errorf("failed to create search task handler: %v", err)
	}

	searchError, err := query.NewSearchErrorTaskHandler(db, cluster, ring, config.Cluster.Current, servers)
	if err != nil {
		return a, fmt.Errorf("failed to create search error task handler: %v", err)
	}

	return app.Application{
		Commands: app.Commands{
			AddTask:               addTask,
			UpdateTask:            updateTask,
			OwnerReg:              ownerReg,
			OwnerUnReg:            ownerUnReg,
			SearchDeleteTask:      searchDeleteTask,
			SearchDeleteErrorTask: searchDeleteErrorTask,
			SearchUpdateTask:      searchUpdateTask,
			SearchUpdateErrorTask: searchUpdateErrorTask,
			HealthCheck:           healthCheck,
		},
		Queries: app.Queries{
			GetFirstInGroup: getFirstInGroup,
			Get:             get,
			Pool:            pool,
			SearchTask:      searchTask,
			SearchError:     searchError,
		},
	}, nil
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
