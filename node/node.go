package main

import (
	"encoding/json"
	"flag"
	"net"
	"syscall"
	"time"

	"code.google.com/p/gcfg"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/jcelliott/lumber"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
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

type mgmNode struct {
	logger core.Logger
}

func main() {
	n := mgmNode{lumber.NewConsoleLogger(lumber.DEBUG)}

	cfgPtr := flag.String("config", "/opt/mgm/node.gcfg", "path to config file")
	flag.Parse()

	//read configuration file
	config := nodeConfig{}
	err := gcfg.ReadFileInto(&config, *cfgPtr)
	if err != nil {
		n.logger.Fatal("Error reading config file: ", err)
		return
	}

	n.logger.Info("config loaded: ", config)

	hStats := make(chan core.HostStats, 8)
	mgmCommands := make(chan []byte, 32)
	socketClosed := make(chan bool)

	go n.collectHostStatistics(hStats)

	for {
		conn, err := net.Dial("tcp", config.Node.MGMAddress)
		if err != nil {
			n.logger.Fatal("Cannot connect to MGM: ", err)
			time.Sleep(10 * time.Second)
			continue
		}
		n.logger.Info("MGM Node connected to MGM")
		go n.readConnection(conn, mgmCommands, socketClosed)

	ProcessingPackets:
		for {
			select {
			case <-socketClosed:
				break ProcessingPackets
			case msg := <-mgmCommands:
				n.logger.Info("recieved message from MGM: ", string(msg))
			case stats := <-hStats:
				nmsg := core.NetworkMessage{}
				nmsg.MessageType = "host_stats"
				nmsg.HStats = stats
				data, err := json.Marshal(nmsg)
				if err != nil {
					n.logger.Error("Error json marshalling stats object: ", err)
					continue
				}
				_, err = conn.Write(data)
				if err == syscall.EPIPE {
					break
				}
				if err != nil {
					n.logger.Error("Error sending data: ", err)
				}
			}
		}
	}

}

func (node mgmNode) readConnection(conn net.Conn, out chan []byte, closing chan bool) {
	for {
		data := make([]byte, 512)
		_, err := conn.Read(data)
		if err != nil {
			node.logger.Error("Error reading from socket: ", err)
			closing <- true
			return
		}
		out <- data
	}
}

func (node mgmNode) collectHostStatistics(out chan core.HostStats) {
	for {
		s := core.HostStats{}
		c, err := cpu.CPUPercent(time.Second, true)
		if err != nil {
			node.logger.Error("Error readin CPU: ", err)
		}
		s.CPUPercent = c

		v, err := mem.VirtualMemory()
		if err != nil {
			node.logger.Error("Error reading Memory", err)
		}
		s.MEMTotal = v.Total
		s.MEMUsed = v.Used
		s.MEMPercent = v.UsedPercent

		out <- s
	}
}