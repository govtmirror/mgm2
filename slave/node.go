package main

import (
	"flag"
	"net"
	"time"

	"code.google.com/p/gcfg"

	"github.com/jcelliott/lumber"
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/remote"
	"github.com/satori/go.uuid"
	pscpu "github.com/shirou/gopsutil/cpu"
	psmem "github.com/shirou/gopsutil/mem"
	psnet "github.com/shirou/gopsutil/net"
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
	connectedAtLeastOnce := false

	cfgPtr := flag.String("config", "/opt/mgm/node.gcfg", "path to config file")
	flag.Parse()

	//read configuration file
	config := nodeConfig{}
	err := gcfg.ReadFileInto(&config, *cfgPtr)
	if err != nil {
		n.logger.Fatal("Error reading config file: ", err)
		return
	}

	n.logger.Info("config loaded successfully")
	regions := map[uuid.UUID]remote.Region{}

	hStats := make(chan mgm.HostStat, 8)
	go n.collectHostStatistics(hStats)

	rMgr := remote.NewRegionManager(config.Node.OpensimBinDir, config.Node.RegionDir, n.logger)
	err = rMgr.Initialize()
	if err != nil {
		n.logger.Error("Error instantiating RegionManager: ", err.Error())
		return
	}

	for {
		n.logger.Info("Connecting to MGM")
		conn, err := net.Dial("tcp", config.Node.MGMAddress)
		if err != nil {
			n.logger.Fatal("Cannot connect to MGM")
			time.Sleep(10 * time.Second)
			continue
		}
		n.logger.Info("MGM Node connected to MGM")

		socketClosed := make(chan bool)
		receiveChan := make(chan core.NetworkMessage, 32)
		sendChan := make(chan core.NetworkMessage, 32)
		nc := host.HostComms{
			Connection: conn,
			Closing:    make(chan bool),
			Log:        n.logger,
		}
		go nc.ReadConnection(receiveChan)
		go nc.WriteConnection(sendChan)

		if !connectedAtLeastOnce {
			//new connection, check for region changes since startup
			sendChan <- core.NetworkMessage{MessageType: "GetRegions"}

			connectedAtLeastOnce = true
		}

	ProcessingPackets:
		for {
			select {
			case <-socketClosed:
				n.logger.Error("Disconnected from MGM")
				time.Sleep(10 * time.Second)
				break ProcessingPackets
			case stats := <-hStats:
				nmsg := core.NetworkMessage{}
				nmsg.MessageType = "HostStats"
				nmsg.HStats = stats
				sendChan <- nmsg
			case msg := <-receiveChan:
				switch msg.MessageType {
				case "AddRegion":
					r := msg.Region
					reg, err := rMgr.AddRegion(r)
					regions[r.UUID] = reg
					if err != nil {
						n.logger.Error("Error adding region: ", err.Error())
					}
					n.logger.Info("AddRegion: %v Complete", r.UUID.String())
				default:
					n.logger.Info("unexpected message from MGM: %v", msg.MessageType)
				}

			}
		}
	}

}

func (node mgmNode) collectHostStatistics(out chan mgm.HostStat) {
	for {
		//start calculating network sent
		fInet, err := psnet.NetIOCounters(false)
		if err != nil {
			node.logger.Error("Error reading networking", err)
		}

		s := mgm.HostStat{}
		c, err := pscpu.CPUPercent(time.Second, true)
		if err != nil {
			node.logger.Error("Error readin CPU: ", err)
		}
		s.CPUPercent = c

		v, err := psmem.VirtualMemory()
		if err != nil {
			node.logger.Error("Error reading Memory", err)
		}
		s.MEMTotal = v.Total / 1000
		s.MEMUsed = (v.Total - v.Available) / 1000
		s.MEMPercent = v.UsedPercent

		lInet, err := psnet.NetIOCounters(false)
		if err != nil {
			node.logger.Error("Error reading networking", err)
		}
		s.NetSent = (lInet[0].BytesSent - fInet[0].BytesSent)
		s.NetRecv = (lInet[0].BytesRecv - fInet[0].BytesRecv)

		out <- s
	}
}
