package host

import (
	"net"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/mgm"
)

type nodeSession struct {
	host         mgm.Host
	conn         net.Conn
	hostSubs     core.SubscriptionManager
	hostStatSubs core.SubscriptionManager
	regionMgr    regionManager
	nodeMgr      nm
	log          core.Logger
}

func (ns nodeSession) process() {
	//place host online
	ns.host.Running = true
	ns.hostSubs.Broadcast(ns.host)

	readMsgs := make(chan Message, 32)
	writeMsgs := make(chan Message, 32)
	nc := Comms{
		Connection: ns.conn,
		Closing:    make(chan bool),
		Log:        ns.log,
	}
	go nc.ReadConnection(readMsgs)
	go nc.WriteConnection(writeMsgs)

	for {

		select {
		case <-nc.Closing:
			ns.log.Info("mgm node disconnected")
			ns.host.Running = false
			ns.hostSubs.Broadcast(ns.host)
			return
		case nmsg := <-readMsgs:
			switch nmsg.MessageType {
			case "Register":
				reg := nmsg.Register
				h, err := ns.nodeMgr.db.UpdateHost(ns.host, reg)
				ns.host = h
				if err != nil {
					ns.log.Error("Error registering new host: ", err.Error())
				}
				ns.hostSubs.Broadcast(ns.host)
			case "HostStats":
				hStats := nmsg.HStats
				hStats.ID = ns.host.ID
				ns.hostStatSubs.Broadcast(hStats)
			case "GetRegions":
				ns.log.Info("Host %v requesting regions list: ", ns.host.ID)
				regions, err := ns.regionMgr.GetRegionsOnHost(ns.host)
				if err != nil {
					ns.log.Error("Error getting regions for host: ", err.Error())
				} else {
					ns.log.Info("Serving %v regions to Host %v", len(regions), ns.host.ID)
					for _, r := range regions {
						writeMsgs <- Message{MessageType: "AddRegion", Region: r}
					}
				}
				ns.log.Info("Region list served to Host %v", ns.host.ID)
			default:
				ns.log.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
			}
		}

	}

}
