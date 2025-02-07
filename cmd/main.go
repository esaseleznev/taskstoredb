package main

import (
	"log"
	"os"

	"github.com/esaseleznev/taskstoredb/internal/adapters"
	"github.com/esaseleznev/taskstoredb/internal/app"
	"github.com/esaseleznev/taskstoredb/internal/app/command"
	"github.com/esaseleznev/taskstoredb/internal/config"
	ports "github.com/esaseleznev/taskstoredb/internal/ports/http"
	"github.com/serialx/hashring"
	"github.com/syndtr/goleveldb/leveldb"
)

// func handler(w http.ResponseWriter, r *http.Request) {
// 	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
// 	logger.Printf("Received request: %s %s, from %s", r.Method, r.URL.Path, r.RemoteAddr)
// 	fmt.Fprintf(w, "Hello, World!")
// }

func main() {

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	config := config.NewConfig()
	application := newApplication(config, logger)
	httpServer := ports.NewHttpServer(config.Http.Port, application, logger)
	httpServer.Start()
}

func newApplication( /*ctx context.Context,*/ config config.Config, logger *log.Logger) app.Application {
	db, err := leveldb.OpenFile(config.Db.Path, nil)
	if err != nil {
		logger.Fatalf("Could not open leveldb %+v\n", err)
	}
	r := adapters.NewLevelRepository(db)

	servers := config.Cluster.Servers
	_ = hashring.New(servers)

	return app.Application{
		Commands: app.Commands{
			AddTask:    command.NewAddTaskHandler(r),
			UpdateTask: command.NewUpdateTaskHendler(r),
		},
	}

}
