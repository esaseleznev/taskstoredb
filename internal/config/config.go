package config

import (
	"flag"
	"log"
	"os"
	"strings"
)

type Config struct {
	Db struct {
		Path string
		Kind string
	}

	Http struct {
		Port string
	}

	Cluster struct {
		Servers []string
		Current string
	}
}

func NewConfig(logger *log.Logger) (Config, error) {
	config := Config{}
	pathDb := flag.String("pdb", "", "path db")
	kindDb := flag.String("kdb", "", "kind db")
	port := flag.String("sp", "", "http port")
	protocol := flag.String("protocol", "", "http or https")
	servers := flag.String("srvs", "", "cluster servers")
	flag.Parse()

	if *pathDb == "" {
		if *pathDb = os.Getenv("TSB_PDB"); *pathDb == "" {
			logger.Println("Path to db not specified, use current directory")
			currentDir, err := os.Getwd()
			if err != nil {
				return config, err
			}
			*pathDb = currentDir + "/data"

		}
	}
	if *kindDb == "" {
		if *kindDb = os.Getenv("TSB_KDB"); *kindDb == "" {
			logger.Println("Kind of db not specified, use leveldb")
			*kindDb = "leveldb"
		}
	}
	config.Db.Kind = *kindDb

	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	if *port == "" {
		if *port = os.Getenv("TSB_SP"); *port == "" {
			logger.Println("Http port not specified, use default port 8080")
			*port = "8080"
		}
	}
	if *protocol == "" {
		logger.Println("Protocol not specified, use default protocol http")
		*protocol = "http"
	}

	config.Db.Path = *pathDb + "/" + host
	config.Http.Port = *port
	config.Cluster.Current = *protocol + "://" + host + ":" + *port

	if *servers == "" {
		if *servers = os.Getenv("TSB_SRVS"); *servers == "" {
			*servers = config.Cluster.Current
		}
	}
	config.Cluster.Servers = strings.Split(*servers, ",")

	return config, nil
}
