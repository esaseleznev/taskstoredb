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

	Cluster struct {
		Servers     []string
		Current     string
		CurrentPort string
	}

	Raft struct {
		Path    string
		Servers []RaftNode
		Current *RaftNode
	}
}

type RaftNode struct {
	Id      string
	Address string
}

func NewConfig(logger *log.Logger) (Config, error) {
	config := Config{}
	pathDb := flag.String("pdb", "", "path db")
	kindDb := flag.String("kdb", "", "kind db")
	сport := flag.String("cport", "", "http port")
	сservers := flag.String("csrvs", "", "cluster servers")
	caddr := flag.String("caddr", "", "curent cluster server")

	rpath := flag.String("rpath", "", "path db")
	rservers := flag.String("rsrvs", "", "cluster servers")
	raddr := flag.String("raddr", "", "curent cluster server")

	protocol := flag.String("protocol", "", "http or https or other")
	flag.Parse()

	host, err := os.Hostname()
	if err != nil {
		return config, err
	}

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

	if *сport == "" {
		if *сport = os.Getenv("TSB_CPORT"); *сport == "" {
			logger.Println("Http port not specified, use default port 8080")
			*сport = "8080"
		}
	}

	if *protocol == "" {
		logger.Println("Protocol not specified, use default protocol http")
		*protocol = "http"
	}

	if *caddr == "" {
		if *caddr = os.Getenv("TSB_CADDR"); *caddr == "" {
			logger.Println("Current cluster server not specified, use default server")
			*caddr = *protocol + "://" + host + ":" + *сport
		}
	}

	config.Db.Path = *pathDb + "/" + host
	config.Cluster.CurrentPort = *сport
	config.Cluster.Current = *caddr

	if *сservers == "" {
		if *сservers = os.Getenv("TSB_CSRVS"); *сservers == "" {
			*сservers = config.Cluster.Current
		}
	}
	config.Cluster.Servers = strings.Split(*сservers, ",")

	if *rpath == "" {
		if *rpath = os.Getenv("TSB_RPATH"); *rpath == "" {
			logger.Println("Path to raft not specified, use current directory")
			currentDir, err := os.Getwd()
			if err != nil {
				return config, err
			}
			*rpath = currentDir + "/raft"

		}
	}
	config.Raft.Path = *rpath + "/" + host

	if *raddr == "" {
		*raddr = os.Getenv("TSB_RADDR")
	}
	if *raddr != "" {
		it := strings.Split(*raddr, ",")
		if len(it) == 2 {
			config.Raft.Current = &RaftNode{Id: it[0], Address: it[1]}
		}
	}

	if *rservers == "" {
		if *rservers = os.Getenv("TSB_RSRVS"); *rservers == "" {
			rservers = raddr
		}
	}
	if *rservers != "" {
		it := strings.Split(*rservers, ",")
		var id string
		var addres string
		for i, server := range it {
			if i+1%2 == 0 {
				addres = server
			} else {
				id = server
				config.Raft.Servers = append(config.Raft.Servers, RaftNode{Id: id, Address: addres})
			}
		}
	}

	return config, nil
}
