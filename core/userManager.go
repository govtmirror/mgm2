package core

import (
	"fmt"

	"github.com/satori/go.uuid"
)

// UserManager is the process that listens for new session connections and spins of the session go-routine
func UserManager(sessionListener <-chan UserSession, jobNotify <-chan Job, dataStore Database, userConn UserConnector, logger Logger) {

	//structure for our lookup table
	type sessionLookup struct {
		jobLink chan Job
	}

	userMap := make(map[uuid.UUID]sessionLookup)
	clientClosed := make(chan uuid.UUID, 64)

	//create notification hub

	//listen for user sessions and hook them in
	go func() {
		for {
			select {
			case s := <-sessionListener:
				//new user session
				userMap[s.GetGUID()] = sessionLookup{make(chan Job, 32)}
				logger.Info("User %v Connected", s.GetGUID().String())
				go userSession(s, userMap[s.GetGUID()].jobLink, dataStore, userConn, logger)
			case id := <-clientClosed:
				//user session has disconnected
				logger.Info("User %v Disconnected", id.String())
				delete(userMap, id)
			case j := <-jobNotify:
				//a job has been updated, is this user connected?
				session, ok := userMap[j.User]
				if ok {
					//tell session coroutine to inform user
					session.jobLink <- j
				}
			}
		}
	}()

}

func userSession(session UserSession, jobLink <-chan Job, dataStore Database, userConn UserConnector, logger Logger) {

	clientMsg := make(chan []byte, 32)
	clientClosed := make(chan bool)

	go session.Read(clientMsg, clientClosed)

	for {
		select {
		case j := <-jobLink:
			session.SendJob(0, j)
		case msg := <-clientMsg:
			//message from client
			m := userRequest{}
			m.load(msg)
			switch m.MessageType {
			case "IarUpload":
				userID, password, err := m.readPassword()
				if err != nil {
					logger.Error("Error reading iar upload request")
					continue
				}
				logger.Info("Iar upload request from %v:%v", userID, password)
				isValid, err := userConn.ValidatePassword(userID, password)
				if err != nil {
					session.SignalError(m.MessageID, err.Error())
				} else {
					if isValid {
						//password is valid, create the upload job
						job, err := dataStore.CreateLoadIarJob(userID, "/")
						if err != nil {
							session.SignalError(m.MessageID, err.Error())
						} else {
							session.SendJob(m.MessageID, job)
							session.SignalSuccess(m.MessageID, fmt.Sprintf("%v", job.ID))
						}
					} else {
						session.SignalError(m.MessageID, "Invalid Password")
					}
				}
			case "SetPassword":
				userID, password, err := m.readPassword()
				if err != nil {
					logger.Error("Error reading password request")
					continue
				}
				logger.Info("Setting password for %v to %v", userID, password)
				if userID != session.GetGUID() && session.GetAccessLevel() < 250 {
					session.SignalError(m.MessageID, "Permission Denied")
				} else {
					if password == "" {
						session.SignalError(m.MessageID, "Password Cannot be blank")
					} else {
						err = userConn.SetPassword(session.GetGUID(), password)
						if err != nil {
							session.SignalError(m.MessageID, err.Error())
						} else {
							session.SignalSuccess(m.MessageID, "Password Set Successfully")
						}
					}
				}
			case "GetDefaultConfig":
				if session.GetAccessLevel() > 249 {
					logger.Info("Serving Default Region Configs.  Request: %v", m.MessageID)
					cfgs, err := dataStore.GetDefaultConfigs()
					if err != nil {
						logger.Error("Error getting default configs: ", err)
					} else {
						for _, cfg := range cfgs {
							session.SendConfig(m.MessageID, cfg)
						}
						session.SignalSuccess(m.MessageID, "Default Config Retrieved")
					}
				}
			case "GetConfig":
				if session.GetAccessLevel() > 249 {
					logger.Info("Serving Region Configs.  Request: %v", m.MessageID)
					rid, err := m.readRegionID()
					if err != nil {
						logger.Error("Error reading region id for configs: ", err)
					} else {
						logger.Info("Serving Region Configs for %v.", rid)
						cfgs, err := dataStore.GetConfigs(rid)
						if err != nil {
							logger.Error("Error getting configs: ", err)
						} else {
							for _, cfg := range cfgs {
								session.SendConfig(m.MessageID, cfg)
							}
							session.SignalSuccess(m.MessageID, "Config Retrieved")
						}
					}
				}
			case "GetState":
				logger.Info("Service state request")
				users, err := userConn.GetUsers()
				if err != nil {
					logger.Error("Error lookin up activeuser account: ", err)
				}
				for _, user := range users {
					if user.Suspended && session.GetAccessLevel() < 250 {
						continue
					}
					session.SendUser(m.MessageID, user)
				}
				users = nil

				jobs, err := dataStore.GetJobsForUser(session.GetGUID())
				if err != nil {
					logger.Error("Error lookin up tasks: ", err)
				}
				for _, job := range jobs {
					session.SendJob(m.MessageID, job)
				}
				jobs = nil

				pendingUsers, err := dataStore.GetPendingUsers()
				if err != nil {
					logger.Error("Error lookin up pending user account: ", err)
				}
				for _, user := range pendingUsers {
					session.SendPendingUser(m.MessageID, user)
				}
				pendingUsers = nil

				//send regions this user may control
				regions, err := dataStore.GetRegions()
				if err != nil {
					logger.Error("Error lookin up user regions: ", err)
				}
				for _, r := range regions {
					session.SendRegion(0, r)
				}
				regions = nil

				//send Estate, Group, and Host dataManager
				estates, err := dataStore.GetEstates()
				if err != nil {
					logger.Error("Error lookin up estates: ", err)
				}
				for _, e := range estates {
					session.SendEstate(m.MessageID, e)
				}
				estates = nil
				groups, err := userConn.GetGroups()
				if err != nil {
					logger.Error("Error lookin up groups: ", err)
				}
				for _, g := range groups {
					session.SendGroup(m.MessageID, g)
				}
				groups = nil
				//only administrative users need host access
				if session.GetAccessLevel() > 249 {
					hosts, err := dataStore.GetHosts()
					if err != nil {
						logger.Error("Error lookin up hosts: ", err)
					}
					for _, h := range hosts {
						session.SendHost(m.MessageID, h)
					}
				}

				//signal to the client that we have completed initial state sync
				session.SignalSuccess(m.MessageID, "State Sync Complete")
				logger.Info("Sync Complete")

			default:
				logger.Error("Error on message from client: ", m.MessageType)

			}
		case <-clientClosed:
			//the client connection has closed
			logger.Info("Client went away")
			return
		}

	}
}
