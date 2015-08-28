package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"code.google.com/p/gcfg"

	"github.com/gorilla/websocket"
	"github.com/jcelliott/lumber"
	"github.com/satori/go.uuid"

	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/remote"
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
		MinRegionPort   uint
		MaxRegionPort   uint
		MinConsolePort  uint
		MaxConsolePort  uint
		ExternalAddress string
	}
}

type mgmNode struct {
	logger logger.Log
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
		n.logger.Fatal("Error reading config file: ", err.Error())
		return
	}

	hostname, err := os.Hostname()
	if err != nil {
		n.logger.Fatal("Error getting hostname: ", err.Error())
		return
	}

	err = validateConfig(config)
	if err != nil {
		n.logger.Fatal("Error in config file: ", err)
		return
	}

	n.logger.Info("config loaded successfully")
	regions := map[uuid.UUID]remote.Region{}

	hStats := make(chan mgm.HostStat, 8)
	go n.collectHostStatistics(hStats)
	rStats := make(chan mgm.RegionStat, 64)

	rMgr := remote.NewRegionManager(config.Node.OpensimBinDir, config.Node.RegionDir, config.Opensim.ExternalAddress, rStats, n.logger)
	err = rMgr.Initialize()
	if err != nil {
		n.logger.Error("Error instantiating RegionManager: ", err.Error())
		return
	}

	for {
		n.logger.Info("Connecting to MGM")
		//conn, err := net.Dial("tcp", config.Node.MGMAddress)
		conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%v/host/", config.Node.MGMAddress), nil)
		if err != nil {
			n.logger.Fatal(fmt.Sprintf("Cannot connect to MGM %v", err.Error()))
			time.Sleep(10 * time.Second)
			continue
		}
		n.logger.Info("MGM Node connected to MGM")

		receiveChan := make(chan host.Message, 32)
		nc := host.Comms{
			Connection: conn,
			Closing:    make(chan bool),
			Log:        n.logger,
		}
		go func() {
			for {
				msg := host.Message{}
				err = conn.ReadJSON(&msg)
				if err != nil {
					nc.Closing <- true
					return
				}
				receiveChan <- msg
			}
		}()

		if !connectedAtLeastOnce {
			//new connection
			//update registration
			reg := host.Registration{}
			reg.ExternalAddress = config.Opensim.ExternalAddress
			reg.Name = hostname
			reg.Slots = int(config.Opensim.MaxRegionPort-config.Opensim.MinRegionPort) + 1
			conn.WriteJSON(host.Message{MessageType: "Register", Register: reg})
			//check for region changes since startup
			conn.WriteJSON(host.Message{MessageType: "GetRegions"})

			connectedAtLeastOnce = true
		}

	ProcessingPackets:
		for {
			select {
			case <-nc.Closing:
				n.logger.Error("Disconnected from MGM")
				time.Sleep(10 * time.Second)
				break ProcessingPackets
			case stats := <-hStats:
				nmsg := host.Message{}
				nmsg.MessageType = "HostStats"
				nmsg.HStats = stats
				conn.WriteJSON(nmsg)
			case stats := <-rStats:
				nmsg := host.Message{}
				nmsg.MessageType = "RegionStats"
				nmsg.RStats = stats
				conn.WriteJSON(nmsg)
			case msg := <-receiveChan:
				switch msg.MessageType {
				case "AddRegion":
					r := msg.Region
					n.logger.Info("AddRegion: %v", r.UUID.String())
					m := host.Message{}

					_, ok := regions[r.UUID]
					if ok {
						//region is already on this node
						m.MessageType = "Success"
						m.Message = "Region added"
					} else {
						//new-to-us region
						reg, err := rMgr.AddRegion(r.UUID)
						regions[r.UUID] = reg
						if err != nil {
							n.logger.Error("Error adding region: ", err.Error())
							m.MessageType = "Failure"
							m.Message = err.Error()
						} else {
							m.MessageType = "Success"
							m.Message = "Region added"
						}
					}
					conn.WriteJSON(m)
					n.logger.Info("AddRegion: %v Complete", r.UUID.String())
				case "RemoveRegion":
					r := msg.Region
					n.logger.Info("RemoveRegion: %v", r.UUID.String())
					m := host.Message{}

					if _, ok := regions[r.UUID]; ok {

						err := rMgr.RemoveRegion(r.UUID)
						if err != nil {
							m.MessageType = "Failure"
							m.Message = err.Error()
							conn.WriteJSON(m)
							continue
						}
						delete(regions, r.UUID)
						m.MessageType = "Success"
						m.Message = "Region removed"
						conn.WriteJSON(m)
						n.logger.Info("RemoveRegion: %v Complete", r.UUID.String())
					} else {
						n.logger.Info("RemoveRegion: %v failed, not present", r.UUID.String())
					}
				case "StartRegion":
					reg := msg.Region
					if r, ok := regions[reg.UUID]; ok {
						//ready response
						m := host.Message{}
						m.ID = msg.ID
						err := r.WriteRegionINI(reg)
						if err != nil {
							errMsg := fmt.Sprintf("Error writing region ini: %v", err.Error())
							n.logger.Error(errMsg)
							m.MessageType = "Failure"
							m.Message = errMsg
							conn.WriteJSON(m)
							continue
						}
						err = r.WriteOpensimINI(msg.Configs)
						if err != nil {
							errMsg := fmt.Sprintf("Error writing opensim ini: %v", err.Error())
							n.logger.Error(errMsg)
							m.MessageType = "Failure"
							m.Message = errMsg
							conn.WriteJSON(m)
							continue
						}
						r.Start()
						m.MessageType = "Success"
						m.Message = "Region started"
						conn.WriteJSON(m)
					}
				case "KillRegion":
					reg := msg.Region
					if r, ok := regions[reg.UUID]; ok {
						//ready response
						m := host.Message{}
						m.ID = msg.ID
						r.Kill()
						m.MessageType = "Success"
						m.Message = "Region killed"
						conn.WriteJSON(m)
					}
				case "RemoveHost":
					n.logger.Info("Received RemoveHost command from MGM, terminating")
					//terminate connection to MGM
					conn.Close()
					//kill any running regions
					//regions are child processes, and are killed as long as we dont crash out
					//exit
					os.Exit(0)
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
		s.Running = true
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

func validateConfig(config nodeConfig) error {
	exists, err := fileExists(config.Node.OpensimBinDir)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("Opensim Bin Dir does not exist")
	}
	exists, err = fileExists(config.Node.RegionDir)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("Region Dir does not exist")
	}
	//skipping ip/hostname validation for now.  Just make sure they arent blank
	if config.Node.MGMAddress == "" {
		return errors.New("MGM address is required")
	}
	if config.Opensim.ExternalAddress == "" {
		return errors.New("External address is required")
	}
	if config.Opensim.MinRegionPort <= 0 || config.Opensim.MinRegionPort > config.Opensim.MaxRegionPort {
		return errors.New("Min Region port must be larger than zero and smaller [or equal to] the Max Region Port")
	}
	if config.Opensim.MaxRegionPort <= 0 {
		return errors.New("Max Region port must be larger than zero")
	}
	if config.Opensim.MinConsolePort <= 0 || config.Opensim.MinConsolePort > config.Opensim.MaxConsolePort {
		return errors.New("Min Console port must be larger than zero and smaller [or equal to] the Max Console Port")
	}
	if config.Opensim.MaxConsolePort <= 0 {
		return errors.New("Max Region port must be larger than zero")
	}
	regionPortSpan := config.Opensim.MaxRegionPort - config.Opensim.MinRegionPort
	consolePortSpan := config.Opensim.MaxConsolePort - config.Opensim.MinConsolePort
	if regionPortSpan != consolePortSpan {
		return errors.New("Regions and consoles should ahve the same number of available ports")
	}
	return nil
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
