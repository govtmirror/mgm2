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
func NewManager(sessionListener <-chan core.UserSession, pers persist.MGMDB, userMgr user.Manager, jobMgr job.Manager, hostMgr host.Manager, regionMgr region.Manager, uConn core.UserConnector, log logger.Log, not Notifier) Manager {
	sMgr := sessionMgr{}
	sMgr.jobMgr = jobMgr
	sMgr.mgm = pers
	sMgr.hostMgr = hostMgr
	sMgr.regionMgr = regionMgr
	sMgr.log = logger.Wrap("SESSION", log)
	sMgr.userConn = uConn
	sMgr.userMgr = userMgr
	sMgr.sessionListener = sessionListener
	sMgr.notifier = not

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
	notifier        Notifier
}

func (sm sessionMgr) process() {

	//listen for user sessions and hook them in
	go func() {

		userMap := make(map[uuid.UUID]Notifier)
		clientClosed := make(chan uuid.UUID, 64)

		for {
			select {
			//FAN OUT MGM EVENTS FOR SESSIONS
			case h := <-sm.notifier.hUp:
				for _, note := range userMap {
					note.HostUpdated(h)
				}
			case h := <-sm.notifier.hDel:
				for _, note := range userMap {
					note.HostDeleted(h)
				}
			case s := <-sm.notifier.hStat:
				for _, note := range userMap {
					note.HostStat(s)
				}
			case r := <-sm.notifier.rUp:
				for _, note := range userMap {
					note.RegionUpdated(r)
				}
			case r := <-sm.notifier.rDel:
				for _, note := range userMap {
					note.RegionDeleted(r)
				}
			case s := <-sm.notifier.rStat:
				for _, note := range userMap {
					note.RegionStat(s)
				}
			case e := <-sm.notifier.eUp:
				for _, note := range userMap {
					note.EstateUpdated(e)
				}
			case e := <-sm.notifier.eDel:
				for _, note := range userMap {
					note.EstateDeleted(e)
				}
			case j := <-sm.notifier.jUp:
				note, ok := userMap[j.User]
				if ok {
					note.JobUpdated(j)
				}
			case j := <-sm.notifier.jDel:
				note, ok := userMap[j.User]
				if ok {
					note.JobDeleted(j)
				}
			// SESSION FUNCTIONS
			case s := <-sm.sessionListener:
				//new user session
				sm.log.Info("User %v Connected", s.GetGUID().String())
				us := userSession{
					client:   s,
					closing:  clientClosed,
					log:      logger.Wrap(s.GetGUID().String(), sm.log),
					mgm:      sm.mgm,
					hMgr:     sm.hostMgr,
					jMgr:     sm.jobMgr,
					notifier: NewNotifier(),
				}
				userMap[s.GetGUID()] = us.notifier
				go us.process()
			case id := <-clientClosed:
				//user session has disconnected
				sm.log.Info("User %v Disconnected", id.String())
				delete(userMap, id)
			}
		}
	}()

}
