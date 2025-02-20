package config

import (
	"flag"
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

func NewConfig() Config {
	config := Config{}
	pathDb := flag.String("pdb", "", "path db")
	kindDb := flag.String("kdb", "", "kind db")
	port := flag.String("sp", "", "http port")
	protocol := flag.String("protocol", "", "http or https")
	servers := flag.String("srvs", "", "cluster servers")
	flag.Parse()

	if *pathDb == "" {
		if *pathDb = os.Getenv("TSB_PDB"); *pathDb == "" {
			panic("Path to db not specified")
		}
	}
	if *kindDb == "" {
		if *kindDb = os.Getenv("TSB_KDB"); *kindDb == "" {
			panic("Path to db not specified")
		}
	}
	config.Db.Kind = *kindDb

	host, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	if *port == "" {
		if *port = os.Getenv("TSB_SP"); *port == "" {
			panic("Http port not specified")
		}
	}
	if *protocol == "" {
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

	return config
}
