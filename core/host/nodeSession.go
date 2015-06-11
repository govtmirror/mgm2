package host

import (
	"fmt"
	"net"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type nodeSession struct {
	host           mgm.Host
	Running        bool
	conn           net.Conn
	hostSubs       core.SubscriptionManager
	hostStatSubs   core.SubscriptionManager
	regionStatSubs core.SubscriptionManager
	regionMgr      region.Manager
	nodeMgr        nm
	cmdMsgs        chan Message
	log            logger.Log
}

func (ns nodeSession) process(closing chan<- int) {
	readMsgs := make(chan Message, 32)
	writeMsgs := make(chan Message, 32)
	nc := Comms{
		Connection: ns.conn,
		Closing:    make(chan bool),
		Log:        ns.log,
	}
	go nc.ReadConnection(readMsgs)
	go nc.WriteConnection(writeMsgs)

	defer ns.conn.Close()

	//place host online
	ns.host.Running = true
	ns.hostSubs.Broadcast(ns.host)

	//track latest region stats, so we can offline them if the node disconnects
	regions := make(map[uuid.UUID]mgm.RegionStat)

	//prepare for request tracking, so we might report results back to users
	var requestNum uint
	pendingRequests := make(map[uint]Message)

	for {

		select {
		case <-nc.Closing:
			ns.log.Info("disconnected")
			//update host broadcasters
			ns.host.Running = false
			ns.hostSubs.Broadcast(ns.host)
			//update region broadcasters
			for _, stat := range regions {
				if stat.Running {
					stat.Running = false
					ns.regionStatSubs.Broadcast(stat)
				}
			}
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

		case nmsg := <-readMsgs:
			// Messages coming from the host
			switch nmsg.MessageType {
			case "Register":
				reg := nmsg.Register
				h, err := ns.nodeMgr.db.UpdateHost(ns.host, reg)
				ns.host = h
				if err != nil {
					ns.log.Error("Error registering: ", err.Error())
				}
				ns.hostSubs.Broadcast(ns.host)
			case "HostStats":
				hStats := nmsg.HStats
				hStats.ID = ns.host.ID
				ns.hostStatSubs.Broadcast(hStats)
			case "RegionStats":
				rStats := nmsg.RStats
				//track stats value
				regions[rStats.UUID] = rStats
				ns.regionStatSubs.Broadcast(rStats)
			case "GetRegions":
				ns.log.Info("requesting regions list")
				regions, err := ns.regionMgr.GetRegionsOnHost(ns.host)
				if err != nil {
					ns.log.Error("Error getting regions: ", err.Error())
				} else {
					ns.log.Info("Serving %v regions", len(regions))
					for _, r := range regions {
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
