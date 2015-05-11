package main

import (
	"flag"

	"code.google.com/p/gcfg"

	"github.com/jcelliott/lumber"
)

type nodeConfig struct {
	Node struct {
		OpensimBinDir string
		RegionDir     string
		MGMAddress    string
	}

	Opensim struct {
		MinRegionPort   int
		MaxRegionPort   int
		MinConsolePort  int
		MaxConsolePort  int
		ExternalAddress string
	}
}

func main() {
	logger := lumber.NewConsoleLogger(lumber.DEBUG)

	cfgPtr := flag.String("config", "/opt/mgm/node.gcfg", "path to config file")

	flag.Parse()

	//read configuration file
	config := nodeConfig{}
	err := gcfg.ReadFileInto(&config, *cfgPtr)
	if err != nil {
		logger.Fatal("Error reading config file: ", err)
		return
	}

	logger.Info("config loaded: ", config)
}
