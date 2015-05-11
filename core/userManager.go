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
				go userSession(s, userMap[s.GetGUID()].jobLink, clientClosed, dataStore, userConn, logger)
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

func userSession(session UserSession, jobLink <-chan Job, exitLink chan<- uuid.UUID, dataStore Database, userConn UserConnector, logger Logger) {

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
			case "DeleteJob":
				logger.Info("User %v requesting delete job", session.GetGUID())
				id, err := m.readID()
				if err != nil {
					session.SignalError(m.MessageID, "Invalid format")
					continue
				}
				job, err := dataStore.GetJobByID(id)
				if err != nil {
					session.SignalError(m.MessageID, "Error retrieving job")
					continue
				}
				if job.ID != id {
					session.SignalError(m.MessageID, "Job does not exist")
					continue
				}
				err = dataStore.DeleteJob(job)
				if err != nil {
					logger.Error("Error deleting job: ", err)
					session.SignalError(m.MessageID, "Error deleting job")
					continue
				}
				session.SignalSuccess(m.MessageID, "Job Deleted")
			case "IarUpload":
				logger.Info("User %v requesting iar upload", session.GetGUID())
				userID, password, err := m.readPassword()
				if err != nil {
					logger.Error("Error reading iar upload request")
					continue
				}
				isValid, err := userConn.ValidatePassword(userID, password)
				if err != nil {
					session.SignalError(m.MessageID, err.Error())
				} else {
					if isValid {
						//password is valid, create the upload job
						job, err := dataStore.CreateLoadIarJob(userID, "/")
						if err != nil {
							logger.Error("Cannot creat job for load_iar: ", err)
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
				logger.Info("User %v requesting password change", session.GetGUID())
				userID, password, err := m.readPassword()
				if err != nil {
					logger.Error("Error reading password request")
					continue
				}
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
							logger.Info("User %v password changed", session.GetGUID())
						}
					}
				}
			case "GetDefaultConfig":
				logger.Info("User %v requesting default configuration", session.GetGUID())
				if session.GetAccessLevel() > 249 {
					cfgs, err := dataStore.GetDefaultConfigs()
					if err != nil {
						logger.Error("Error getting default configs: ", err)
					} else {
						for _, cfg := range cfgs {
							session.SendConfig(m.MessageID, cfg)
						}
						session.SignalSuccess(m.MessageID, "Default Config Retrieved")
						logger.Info("User %v default configuration served", session.GetGUID())
					}
				} else {
					logger.Info("User %v permission denied to default configurations", session.GetGUID())
					session.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetConfig":
				logger.Info("User %v requesting region configuration", session.GetGUID())
				if session.GetAccessLevel() > 249 {
					rid, err := m.readRegionID()
					if err != nil {
						logger.Error("Error reading region id for configs: ", err)
						session.SignalError(m.MessageID, "Error loading region")
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
							logger.Info("User %v config retrieved", session.GetGUID())
						}
					}
				} else {
					logger.Info("User %v permission denied to configurations", session.GetGUID())
					session.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetState":
				logger.Info("User %v requesting state sync", session.GetGUID())
				users, err := userConn.GetUsers()
				if err != nil {
					logger.Error("Error lookin up activeuser account: ", err)
					session.SignalError(m.MessageID, "Error loading user accounts")
					continue
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
					session.SignalError(m.MessageID, "Error loading tasks")
					continue
				}
				for _, job := range jobs {
					session.SendJob(m.MessageID, job)
				}
				jobs = nil

				pendingUsers, err := dataStore.GetPendingUsers()
				if err != nil {
					logger.Error("Error lookin up pending user account: ", err)
					session.SignalError(m.MessageID, "Error looking up pending users")
					continue
				}
				for _, user := range pendingUsers {
					session.SendPendingUser(m.MessageID, user)
				}
				pendingUsers = nil

				//send regions this user may control
				regions, err := dataStore.GetRegions()
				if err != nil {
					logger.Error("Error lookin up user regions: ", err)
					session.SignalError(m.MessageID, "Error looking up regions")
					continue
				}
				for _, r := range regions {
					session.SendRegion(0, r)
				}
				regions = nil

				//send Estate, Group, and Host dataManager
				estates, err := dataStore.GetEstates()
				if err != nil {
					logger.Error("Error lookin up estates: ", err)
					session.SignalError(m.MessageID, "Error looking up estates")
					continue
				}
				for _, e := range estates {
					session.SendEstate(m.MessageID, e)
				}
				estates = nil
				groups, err := userConn.GetGroups()
				if err != nil {
					logger.Error("Error lookin up groups: ", err)
					session.SignalError(m.MessageID, "Error looking up groups")
					continue
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
						session.SignalError(m.MessageID, "Error enumerating hosts")
						continue
					}
					for _, h := range hosts {
						session.SendHost(m.MessageID, h)
					}
				}

				logger.Info("User %v state sync complete", session.GetGUID())
				//signal to the client that we have completed initial state sync
				session.SignalSuccess(m.MessageID, "State Sync Complete")

			default:
				logger.Error("Error on message from client: ", m.MessageType)
				session.SignalError(m.MessageID, "Invalid request")
			}
		case <-clientClosed:
			//the client connection has closed
			exitLink <- session.GetGUID()
			return
		}

	}
}
