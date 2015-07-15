package host

import (
	"fmt"
	"net"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
)

type hostSession struct {
	host    mgm.Host
	Running bool
	conn    net.Conn
	cmdMsgs chan Message
	log     logger.Log

	regions  []mgm.Region
	rCmdChan chan regionCommand
}

func (ns hostSession) process(closing chan<- int64, register chan<- registrationRequest, hStatChan chan<- mgm.HostStat, rStatChan chan<- mgm.RegionStat) {
	readMsgs := make(chan Message, 32)
	writeMsgs := make(chan Message, 32)
	nc := Comms{
		Connection: ns.conn,
		Closing:    make(chan bool),
		Log:        ns.log,
	}
	go nc.ReadConnection(readMsgs)
	go nc.WriteConnection(writeMsgs)

	var regions []mgm.Region

	defer ns.conn.Close()

	//prepare for request tracking, so we might report results back to users
	var requestNum uint
	pendingRequests := make(map[uint]Message)

Processing:
	for {

		select {
		case <-nc.Closing:
			ns.log.Info("disconnected")
			//notify manager that we disconnected
			closing <- ns.host.ID
			return

		case msg := <-ns.cmdMsgs:
			// Messages coming from MGM
			// confirm we are not pending on an identical request
			for _, req := range pendingRequests {
				if req.MessageType == msg.MessageType && req.Region.UUID == msg.Region.UUID {
					msg.SR(false, fmt.Sprintf("Pending operation of type %v already in progress", req.MessageType))
					ns.log.Info("Ignoring request of type %v, matching request already in progress", req.MessageType)
					continue
				}
			}
			//no pending detected, pass it through
			msg.ID = requestNum
			pendingRequests[msg.ID] = msg
			writeMsgs <- msg

		case rCmd := <-ns.rCmdChan:
			switch rCmd.cmd {
			case "AddRegion":
				for _, r := range regions {
					if r.UUID == rCmd.region.UUID {
						//region already on host
						break Processing
					}
				}
				//add region to list
				regions = append(regions, rCmd.region)
				if ns.Running {
					//tell the remote node about it too
					writeMsgs <- Message{MessageType: "AddRegion", Region: rCmd.region}
				}
			case "RemoveRegion":
				for i, r := range regions {
					if r.UUID == rCmd.region.UUID {
						regions = append(regions[:i], regions[i+1:]...)
						if ns.Running {
							//tell the remote node about it too
							writeMsgs <- Message{MessageType: "RemoveRegion", Region: rCmd.region}
						}
					}
				}

			}
		case nmsg := <-readMsgs:
			// Messages coming from the host
			switch nmsg.MessageType {

			case "Register":
				reg := nmsg.Register
				//update existing host
				register <- registrationRequest{reg, ns.host}

			case "HostStats":
				hStats := nmsg.HStats
				hStats.ID = ns.host.ID
				hStatChan <- hStats
			case "RegionStats":
				rStats := nmsg.RStats
				//track stats value
				rStatChan <- rStats
			case "GetRegions":
				ns.log.Info("requesting regions list")
				for _, r := range ns.regions {
					if r.Host == ns.host.ID {
						writeMsgs <- Message{MessageType: "AddRegion", Region: r}
					}
				}
				ns.log.Info("Region list served")
			case "Success":
				//an MGM request has succeeded
				if req, ok := pendingRequests[nmsg.ID]; ok {
					req.SR(true, nmsg.Message)
				}
			case "Failure":
				//an MGM request has failed
				if req, ok := pendingRequests[nmsg.ID]; ok {
					req.SR(false, nmsg.Message)
				}
			default:
				ns.log.Info("Received invalid message: %s", nmsg.MessageType)
			}
		}

	}

}
