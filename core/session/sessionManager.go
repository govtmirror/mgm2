package session

import (
	"fmt"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/core/user"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Manager is the process that listens for new session connections and spins of the session go-routine
type Manager interface {
}

// NewManager constructs a session manager for use
func NewManager(sessionListener <-chan core.UserSession, userMgr user.Manager, jobMgr job.Manager, nodeMgr host.Manager, regionMgr region.Manager, uConn core.UserConnector, log logger.Log) Manager {
	sMgr := sessionMgr{}
	sMgr.jobMgr = jobMgr
	sMgr.nodeMgr = nodeMgr
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
	nodeMgr         host.Manager
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
				userMap[s.GetGUID()] = core.SessionLookup{
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

func (sm sessionMgr) userSession(us core.UserSession, sLinks core.SessionLookup, exitLink chan<- uuid.UUID) {

	clientMsg := make(chan []byte, 32)

	go us.Read(clientMsg)

	host := sm.nodeMgr.SubscribeHost()
	hostStats := sm.nodeMgr.SubscribeHostStats()
	regionStats := sm.nodeMgr.SubscribeRegionStats()

	for {
		select {
		case j := <-sLinks.JobLink:
			us.Send(j)
		case h := <-host.GetReceive():
			if us.GetAccessLevel() > 249 {
				us.Send(h)
			}
		case hs := <-hostStats.GetReceive():
			if us.GetAccessLevel() > 249 {
				us.Send(hs)
			}
		case rs := <-regionStats.GetReceive():
			if us.GetAccessLevel() > 249 {
				us.Send(rs)
			}
		case msg := <-clientMsg:
			//message from client
			m := core.UserRequest{}
			m.Load(msg)
			switch m.MessageType {
			case "StartRegion":
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.SignalError(m.MessageID, "Invalid format")
					continue
				}
				sm.log.Info("User %v requesting start region %v", us.GetGUID(), regionID)
				user, exists, err := sm.userConn.GetUserByID(us.GetGUID())
				if err != nil {
					us.SignalError(m.MessageID, "Error looking up user")
					sm.log.Error("start region %v failed, error finding requesting user", regionID)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, "Invalid requesting user")
					sm.log.Error("start region %v failed, requesting user does not exist", regionID)
					continue
				}
				r, err := sm.regionMgr.GetRegionByID(regionID)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error locating region: %v", err.Error()))
					sm.log.Error("start region %v failed, region not found", regionID)
					continue
				}

				h, err := sm.userMgr.RequestControlPermission(r, user)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error: %v", err.Error()))
					sm.log.Error("start region %v failed: %v", regionID, err.Error())
					continue
				}

				sm.nodeMgr.StartRegionOnHost(r, h, func(success bool, message string) {
					if success {
						us.SignalSuccess(m.MessageID, message)
					} else {
						us.SignalError(m.MessageID, message)
					}
				})
			case "KillRegion":
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.SignalError(m.MessageID, "Invalid format")
					continue
				}
				sm.log.Info("User %v requesting kill region %v", us.GetGUID(), regionID)
				user, exists, err := sm.userConn.GetUserByID(us.GetGUID())
				if err != nil {
					us.SignalError(m.MessageID, "Error looking up user")
					sm.log.Error("kill region %v failed, error finding requesting user", regionID)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, "Invalid requesting user")
					sm.log.Error("kill region %v failed, requesting user does not exist", regionID)
					continue
				}
				r, err := sm.regionMgr.GetRegionByID(regionID)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error locating region: %v", err.Error()))
					sm.log.Error("kill region %v failed, region not found", regionID)
					continue
				}

				h, err := sm.userMgr.RequestControlPermission(r, user)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error requesting permission: %v", err.Error()))
					sm.log.Error("kill region %v failed: %v", regionID, err.Error())
					continue
				}

				sm.nodeMgr.KillRegionOnHost(r, h, func(success bool, message string) {
					if success {
						us.SignalSuccess(m.MessageID, message)
					} else {
						us.SignalError(m.MessageID, message)
					}
				})
			case "DeleteJob":
				sm.log.Info("User %v requesting delete job", us.GetGUID())
				id, err := m.ReadID()
				if err != nil {
					us.SignalError(m.MessageID, "Invalid format")
					continue
				}
				j, err := sm.jobMgr.GetJobByID(id)
				if err != nil {
					us.SignalError(m.MessageID, "Error retrieving job")
					continue
				}
				if j.ID != id {
					us.SignalError(m.MessageID, "Job not found")
					continue
				}
				err = sm.jobMgr.DeleteJob(j)
				if err != nil {
					sm.log.Error("Error deleting job: ", err)
					us.SignalError(m.MessageID, "Error deleting job")
					continue
				}
				//TODO some jobs may need files cleaned up... should we delete them here
				// or leave them and create a cleanup coroutine?
				us.SignalSuccess(m.MessageID, "Job Deleted")
			case "IarUpload":
				sm.log.Info("User %v requesting iar upload", us.GetGUID())
				userID, password, err := m.ReadPassword()
				if err != nil {
					sm.log.Error("Error reading iar upload request")
					continue
				}
				isValid, err := sm.userConn.ValidatePassword(userID, password)
				if err != nil {
					us.SignalError(m.MessageID, err.Error())
				} else {
					if isValid {
						//password is valid, create the upload job
						j, err := sm.jobMgr.CreateLoadIarJob(userID, "/")
						if err != nil {
							sm.log.Error("Cannot creat job for load_iar: ", err)
							us.SignalError(m.MessageID, err.Error())
						} else {
							us.Send(j)
							us.SignalSuccess(m.MessageID, fmt.Sprintf("%v", j.ID))
						}
					} else {
						us.SignalError(m.MessageID, "Invalid Password")
					}
				}
			case "SetPassword":
				sm.log.Info("User %v requesting password change", us.GetGUID())
				userID, password, err := m.ReadPassword()
				if err != nil {
					sm.log.Error("Error reading password request")
					continue
				}
				if userID != us.GetGUID() && us.GetAccessLevel() < 250 {
					us.SignalError(m.MessageID, "Permission Denied")
				} else {
					if password == "" {
						us.SignalError(m.MessageID, "Password Cannot be blank")
					} else {
						err = sm.userConn.SetPassword(us.GetGUID(), password)
						if err != nil {
							us.SignalError(m.MessageID, err.Error())
						} else {
							us.SignalSuccess(m.MessageID, "Password Set Successfully")
							sm.log.Info("User %v password changed", us.GetGUID())
						}
					}
				}
			case "GetDefaultConfig":
				sm.log.Info("User %v requesting default configuration", us.GetGUID())
				if us.GetAccessLevel() > 249 {
					cfgs, err := sm.regionMgr.GetDefaultConfigs()
					if err != nil {
						sm.log.Error("Error getting default configs: ", err)
					} else {
						for _, cfg := range cfgs {
							us.Send(cfg)
						}
						us.SignalSuccess(m.MessageID, "Default Config Retrieved")
						sm.log.Info("User %v default configuration served", us.GetGUID())
					}
				} else {
					sm.log.Info("User %v permission denied to default configurations", us.GetGUID())
					us.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetConfig":
				sm.log.Info("User %v requesting region configuration", us.GetGUID())
				if us.GetAccessLevel() > 249 {
					rid, err := m.ReadRegionID()
					if err != nil {
						sm.log.Error("Error reading region id for configs: ", err)
						us.SignalError(m.MessageID, "Error loading region")
					} else {
						sm.log.Info("Serving Region Configs for %v.", rid)
						cfgs, err := sm.regionMgr.GetConfigs(rid)
						if err != nil {
							sm.log.Error("Error getting configs: ", err)
						} else {
							for _, cfg := range cfgs {
								us.Send(cfg)
							}
							us.SignalSuccess(m.MessageID, "Config Retrieved")
							sm.log.Info("User %v config retrieved", us.GetGUID())
						}
					}
				} else {
					sm.log.Info("User %v permission denied to configurations", us.GetGUID())
					us.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetState":
				sm.log.Info("User %v requesting state sync", us.GetGUID())
				users, err := sm.userConn.GetUsers()
				if err != nil {
					sm.log.Error("Error lookin up activeuser account: ", err)
					us.SignalError(m.MessageID, "Error loading user accounts")
					continue
				}
				for _, user := range users {
					if user.Suspended && us.GetAccessLevel() < 250 {
						continue
					}
					us.Send(user)
				}
				users = nil

				jobs, err := sm.jobMgr.GetJobsForUser(us.GetGUID())
				if err != nil {
					sm.log.Error("Error lookin up tasks: ", err)
					us.SignalError(m.MessageID, "Error loading tasks")
					continue
				}
				for _, j := range jobs {
					us.Send(j)
				}
				jobs = nil

				pendingUsers, err := sm.userMgr.GetPendingUsers()
				if err != nil {
					sm.log.Error("Error lookin up pending user account: ", err)
					us.SignalError(m.MessageID, "Error looking up pending users")
					continue
				}
				for _, user := range pendingUsers {
					us.Send(user)
				}
				pendingUsers = nil

				//send regions this user may control
				regions, err := sm.regionMgr.GetRegions()
				if err != nil {
					sm.log.Error("Error lookin up user regions: ", err)
					us.SignalError(m.MessageID, "Error looking up regions")
					continue
				}
				for _, r := range regions {
					us.Send(r)
				}
				regions = nil

				//send Estate, Group, and Host dataManager
				estates, err := sm.userMgr.GetEstates()
				if err != nil {
					sm.log.Error("Error lookin up estates: ", err)
					us.SignalError(m.MessageID, "Error looking up estates")
					continue
				}
				for _, e := range estates {
					us.Send(e)
				}
				estates = nil

				groups, err := sm.userConn.GetGroups()
				if err != nil {
					sm.log.Error("Error lookin up groups: ", err)
					us.SignalError(m.MessageID, "Error looking up groups")
					continue
				}
				for _, g := range groups {
					us.Send(g)
				}
				groups = nil
				//only administrative users need host access
				if us.GetAccessLevel() > 249 {
					hosts := sm.nodeMgr.GetHosts()
					for _, h := range hosts {
						us.Send(h)
					}
				}

				sm.log.Info("User %v state sync complete", us.GetGUID())
				//signal to the client that we have completed initial state sync
				us.SignalSuccess(m.MessageID, "State Sync Complete")

			default:
				sm.log.Error("Error on message from client: ", m.MessageType)
				us.SignalError(m.MessageID, "Invalid request")
			}
		case <-us.GetClosingSignal():
			//the client connection has closed
			host.Unsubscribe()
			hostStats.Unsubscribe()
			exitLink <- us.GetGUID()
			return
		}

	}
}
