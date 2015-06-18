package session

import (
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/core/user"

	"github.com/satori/go.uuid"
)

// Manager is the process that listens for new session connections and spins of the session go-routine
type Manager interface {
}

// NewManager constructs a session manager for use
func NewManager(sessionListener <-chan core.UserSession, pers persist.MGMDB, userMgr user.Manager, jobMgr job.Manager, hostMgr host.Manager, regionMgr region.Manager, uConn core.UserConnector, log logger.Log) Manager {
	sMgr := sessionMgr{}
	sMgr.jobMgr = jobMgr
	sMgr.mgm = pers
	sMgr.hostMgr = hostMgr
	sMgr.regionMgr = regionMgr
	sMgr.log = logger.Wrap("SESSION", log)
	sMgr.userConn = uConn
	sMgr.userMgr = userMgr
	sMgr.sessionListener = sessionListener

	go sMgr.process()

	return sMgr
}

type sessionMgr struct {
	sessionListener <-chan core.UserSession
	jobMgr          job.Manager
	mgm             persist.MGMDB
	hostMgr         host.Manager
	regionMgr       region.Manager
	userMgr         user.Manager
	userConn        core.UserConnector
	log             logger.Log
}

func (sm sessionMgr) process() {

	userMap := make(map[uuid.UUID]core.SessionLookup)
	clientClosed := make(chan uuid.UUID, 64)

	//listen for user sessions and hook them in
	go func() {
		for {
			select {
			case s := <-sm.sessionListener:
				//new user session
				sm.log.Info("User %v Connected", s.GetGUID().String())
				us := userSession{
					client:  s,
					closing: clientClosed,
					log:     logger.Wrap(s.GetGUID().String(), sm.log),
					mgm:     sm.mgm,
				}
				go us.process()
			case id := <-clientClosed:
				//user session has disconnected
				sm.log.Info("User %v Disconnected", id.String())
				delete(userMap, id)
			}
		}
	}()

}
