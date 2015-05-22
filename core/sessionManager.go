package core

import (
	"fmt"

	"github.com/M-O-S-E-S/mgm/mgm"
	"github.com/satori/go.uuid"
)

// SessionManager is the process that listens for new session connections and spins of the session go-routine
type SessionManager interface {
}

// NewSessionManager constructs a session manager for use
func NewSessionManager(sessionListener <-chan UserSession, jobMgr JobManager, nodeMgr NodeManager, db Database, uConn UserConnector, logger Logger) SessionManager {
	sMgr := sessionMgr{}
	sMgr.jobMgr = jobMgr
	sMgr.nodeMgr = nodeMgr
	sMgr.log = logger
	sMgr.datastore = db
	sMgr.userConn = uConn
	sMgr.sessionListener = sessionListener

	go sMgr.process()

	return sMgr
}

type sessionMgr struct {
	sessionListener <-chan UserSession
	datastore       Database
	jobMgr          JobManager
	nodeMgr         NodeManager
	userConn        UserConnector
	log             Logger
}

func (sm sessionMgr) process() {

	userMap := make(map[uuid.UUID]sessionLookup)
	clientClosed := make(chan uuid.UUID, 64)

	//listen for user sessions and hook them in
	go func() {
		for {
			select {
			case s := <-sm.sessionListener:
				//new user session
				userMap[s.GetGUID()] = sessionLookup{
					make(chan mgm.Job, 32),
					make(chan mgm.HostStat, 32),
					make(chan mgm.Host, 8),
					s.GetAccessLevel(),
				}
				sm.log.Info("User %v Connected", s.GetGUID().String())
				go sm.userSession(s, userMap[s.GetGUID()], clientClosed)
			case id := <-clientClosed:
				//user session has disconnected
				sm.log.Info("User %v Disconnected", id.String())
				delete(userMap, id)
			}
		}
	}()

}

func (sm sessionMgr) userSession(session UserSession, sLinks sessionLookup, exitLink chan<- uuid.UUID) {

	clientMsg := make(chan []byte, 32)

	go session.Read(clientMsg)

	host := sm.nodeMgr.SubscribeHost()
	hostStats := sm.nodeMgr.SubscribeHostStats()

	for {
		select {
		case j := <-sLinks.jobLink:
			session.GetSend() <- j
		case h := <-host.GetReceive():
			if session.GetAccessLevel() > 249 {
				session.GetSend() <- h
			}
		case hs := <-hostStats.GetReceive():
			if session.GetAccessLevel() > 249 {
				session.GetSend() <- hs
			}
		case msg := <-clientMsg:
			//message from client
			m := userRequest{}
			m.load(msg)
			switch m.MessageType {
			case "DeleteJob":
				sm.log.Info("User %v requesting delete job", session.GetGUID())
				id, err := m.readID()
				if err != nil {
					session.SignalError(m.MessageID, "Invalid format")
					continue
				}
				job, err := sm.datastore.GetJobByID(id)
				if err != nil {
					session.SignalError(m.MessageID, "Error retrieving job")
					continue
				}
				if job.ID != id {
					session.SignalError(m.MessageID, "Job not found")
					continue
				}
				err = sm.datastore.DeleteJob(job)
				if err != nil {
					sm.log.Error("Error deleting job: ", err)
					session.SignalError(m.MessageID, "Error deleting job")
					continue
				}
				//TODO some jobs may need files cleaned up... should we delete them here
				// or leave them and create a cleanup coroutine?
				session.SignalSuccess(m.MessageID, "Job Deleted")
			case "IarUpload":
				sm.log.Info("User %v requesting iar upload", session.GetGUID())
				userID, password, err := m.readPassword()
				if err != nil {
					sm.log.Error("Error reading iar upload request")
					continue
				}
				isValid, err := sm.userConn.ValidatePassword(userID, password)
				if err != nil {
					session.SignalError(m.MessageID, err.Error())
				} else {
					if isValid {
						//password is valid, create the upload job
						job, err := sm.datastore.CreateLoadIarJob(userID, "/")
						if err != nil {
							sm.log.Error("Cannot creat job for load_iar: ", err)
							session.SignalError(m.MessageID, err.Error())
						} else {
							session.GetSend() <- job
							session.SignalSuccess(m.MessageID, fmt.Sprintf("%v", job.ID))
						}
					} else {
						session.SignalError(m.MessageID, "Invalid Password")
					}
				}
			case "SetPassword":
				sm.log.Info("User %v requesting password change", session.GetGUID())
				userID, password, err := m.readPassword()
				if err != nil {
					sm.log.Error("Error reading password request")
					continue
				}
				if userID != session.GetGUID() && session.GetAccessLevel() < 250 {
					session.SignalError(m.MessageID, "Permission Denied")
				} else {
					if password == "" {
						session.SignalError(m.MessageID, "Password Cannot be blank")
					} else {
						err = sm.userConn.SetPassword(session.GetGUID(), password)
						if err != nil {
							session.SignalError(m.MessageID, err.Error())
						} else {
							session.SignalSuccess(m.MessageID, "Password Set Successfully")
							sm.log.Info("User %v password changed", session.GetGUID())
						}
					}
				}
			case "GetDefaultConfig":
				sm.log.Info("User %v requesting default configuration", session.GetGUID())
				if session.GetAccessLevel() > 249 {
					cfgs, err := sm.datastore.GetDefaultConfigs()
					if err != nil {
						sm.log.Error("Error getting default configs: ", err)
					} else {
						for _, cfg := range cfgs {
							session.GetSend() <- cfg
						}
						session.SignalSuccess(m.MessageID, "Default Config Retrieved")
						sm.log.Info("User %v default configuration served", session.GetGUID())
					}
				} else {
					sm.log.Info("User %v permission denied to default configurations", session.GetGUID())
					session.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetConfig":
				sm.log.Info("User %v requesting region configuration", session.GetGUID())
				if session.GetAccessLevel() > 249 {
					rid, err := m.readRegionID()
					if err != nil {
						sm.log.Error("Error reading region id for configs: ", err)
						session.SignalError(m.MessageID, "Error loading region")
					} else {
						sm.log.Info("Serving Region Configs for %v.", rid)
						cfgs, err := sm.datastore.GetConfigs(rid)
						if err != nil {
							sm.log.Error("Error getting configs: ", err)
						} else {
							for _, cfg := range cfgs {
								session.GetSend() <- cfg
							}
							session.SignalSuccess(m.MessageID, "Config Retrieved")
							sm.log.Info("User %v config retrieved", session.GetGUID())
						}
					}
				} else {
					sm.log.Info("User %v permission denied to configurations", session.GetGUID())
					session.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetState":
				sm.log.Info("User %v requesting state sync", session.GetGUID())
				users, err := sm.userConn.GetUsers()
				if err != nil {
					sm.log.Error("Error lookin up activeuser account: ", err)
					session.SignalError(m.MessageID, "Error loading user accounts")
					continue
				}
				for _, user := range users {
					if user.Suspended && session.GetAccessLevel() < 250 {
						continue
					}
					session.GetSend() <- user
				}
				users = nil

				jobs, err := sm.datastore.GetJobsForUser(session.GetGUID())
				if err != nil {
					sm.log.Error("Error lookin up tasks: ", err)
					session.SignalError(m.MessageID, "Error loading tasks")
					continue
				}
				for _, job := range jobs {
					session.GetSend() <- job
				}
				jobs = nil

				pendingUsers, err := sm.datastore.GetPendingUsers()
				if err != nil {
					sm.log.Error("Error lookin up pending user account: ", err)
					session.SignalError(m.MessageID, "Error looking up pending users")
					continue
				}
				for _, user := range pendingUsers {
					session.GetSend() <- user
				}
				pendingUsers = nil

				//send regions this user may control
				regions, err := sm.datastore.GetRegions()
				if err != nil {
					sm.log.Error("Error lookin up user regions: ", err)
					session.SignalError(m.MessageID, "Error looking up regions")
					continue
				}
				for _, r := range regions {
					session.GetSend() <- r
				}
				regions = nil

				//send Estate, Group, and Host dataManager
				estates, err := sm.datastore.GetEstates()
				if err != nil {
					sm.log.Error("Error lookin up estates: ", err)
					session.SignalError(m.MessageID, "Error looking up estates")
					continue
				}
				for _, e := range estates {
					session.GetSend() <- e
				}
				estates = nil

				groups, err := sm.userConn.GetGroups()
				if err != nil {
					sm.log.Error("Error lookin up groups: ", err)
					session.SignalError(m.MessageID, "Error looking up groups")
					continue
				}
				for _, g := range groups {
					session.GetSend() <- g
				}
				groups = nil
				//only administrative users need host access
				if session.GetAccessLevel() > 249 {
					hosts, err := sm.datastore.GetHosts()
					if err != nil {
						sm.log.Error("Error lookin up hosts: ", err)
						session.SignalError(m.MessageID, "Error enumerating hosts")
						continue
					}
					for _, h := range hosts {
						session.GetSend() <- h
					}
				}

				sm.log.Info("User %v state sync complete", session.GetGUID())
				//signal to the client that we have completed initial state sync
				session.SignalSuccess(m.MessageID, "State Sync Complete")

			default:
				sm.log.Error("Error on message from client: ", m.MessageType)
				session.SignalError(m.MessageID, "Invalid request")
			}
		case <-session.GetClosingSignal():
			//the client connection has closed
			host.Unsubscribe()
			hostStats.Unsubscribe()
			exitLink <- session.GetGUID()
			return
		}

	}
}
