package session

import (
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type userSession struct {
	client  core.UserSession
	closing chan<- uuid.UUID
	mgm     persist.MGMDB
	log     logger.Log
}

func (us userSession) process() {

	clientMsg := make(chan []byte, 32)

	go us.client.Read(clientMsg)

	//host := sm.hostMgr.SubscribeHost()
	//hostStats := sm.hostMgr.SubscribeHostStats()
	//regionStats := sm.hostMgr.SubscribeRegionStats()

	var console region.RestConsole

	for {
		select {
		//case j := <-sLinks.JobLink:
		//	us.client.Send(j)
		//case h := <-host.GetReceive():
		//	if us.GetAccessLevel() > 249 {
		//		us.Send(h)
		//	}
		//case hs := <-hostStats.GetReceive():
		//	if us.GetAccessLevel() > 249 {
		//		us.Send(hs)
		//	}
		//case rs := <-regionStats.GetReceive():
		//	if us.GetAccessLevel() > 249 {
		//		us.Send(rs)
		//	}
		case msg := <-clientMsg:
			//message from client
			m := core.UserRequest{}
			m.Load(msg)
			switch m.MessageType {
			case "RemoveHost":
				if us.client.GetAccessLevel() < 250 {
					us.client.SignalError(m.MessageID, "Permission Denied")
					continue
				}
				//hostID, err := m.ReadID()
				//if err != nil {
				//	us.SignalError(m.MessageID, "Invalid format")
				//	continue
				//}
				//h, exists, err := sm.hostMgr.GetHostByID(hostID)
				//if err != nil {
				//	us.SignalError(m.MessageID, "Error looking up host")
				//	errMsg := fmt.Sprintf("delete host %v failed, error finding host", hostID)
				//	sm.log.Error(errMsg)
				//	continue
				//}
				//if !exists {
				//	us.SignalError(m.MessageID, "Invalid host")
				//	errMsg := fmt.Sprintf("delete host %v failed, host does not exist", hostID)
				//	sm.log.Error(errMsg)
				//	continue
				//}

				//clean off any regions on the host
				//loop over regions, unassigning them (node kills them if running)
				//for _, uuid := range h.Regions {
				//	r, exists, err := sm.regionMgr.GetRegionByID(uuid)
				//	if err != nil || !exists {
				//		continue
				//	}
				//	sm.regionMgr.SetHostForRegion(r, mgm.Host{})
				//}

				//var wg sync.WaitGroup
				//wg.Add(1)

				//err = sm.hostMgr.RemoveHost(h)
				//if err != nil {
				//	us.SignalError(m.MessageID, err.Error())
				//}
				//hd := mgm.HostRemoved{}
				//hd.ID = h.ID
				//us.Send(hd)
				//us.SignalSuccess(m.MessageID, "host removed")

			case "StartRegion":
				/*regionID, err := m.ReadRegionID()
				if err != nil {
					us.SignalError(m.MessageID, "Invalid format")
					continue
				}
				sm.log.Info("User %v requesting start region %v", us.GetGUID(), regionID)
				user, exists, err := sm.userConn.GetUserByID(us.GetGUID())
				if err != nil {
					us.SignalError(m.MessageID, "Error looking up user")
					errMsg := fmt.Sprintf("start region %v failed, error finding requesting user", regionID)
					sm.log.Error(errMsg)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, "Invalid requesting user")
					errMsg := fmt.Sprintf("start region %v failed, requesting user does not exist", regionID)
					sm.log.Error(errMsg)
					continue
				}
				r, exists, err := sm.regionMgr.GetRegionByID(regionID)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error locating region: %v", err.Error()))
					errMsg := fmt.Sprintf("start region %v failed", regionID)
					sm.log.Error(errMsg)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, fmt.Sprintf("Region does not exist"))
					errMsg := fmt.Sprintf("start region %v failed, region not found", regionID)
					sm.log.Error(errMsg)
					continue
				}

				h, err := sm.userMgr.RequestControlPermission(r, user)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error: %v", err.Error()))
					errMsg := fmt.Sprintf("start region %v failed: %v", regionID, err.Error())
					sm.log.Error(errMsg)
					continue
				}

				sm.hostMgr.StartRegionOnHost(r, h, func(success bool, message string) {
					if success {
						us.SignalSuccess(m.MessageID, message)
					} else {
						us.SignalError(m.MessageID, message)
					}
				})*/
			case "KillRegion":
				/*regionID, err := m.ReadRegionID()
				if err != nil {
					us.SignalError(m.MessageID, "Invalid format")
					continue
				}
				sm.log.Info("User %v requesting kill region %v", us.GetGUID(), regionID)
				user, exists, err := sm.userConn.GetUserByID(us.GetGUID())
				if err != nil {
					us.SignalError(m.MessageID, "Error looking up user")
					errMsg := fmt.Sprintf("kill region %v failed, error finding requesting user", regionID)
					sm.log.Error(errMsg)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, "Invalid requesting user")
					errMsg := fmt.Sprintf("kill region %v failed, requesting user does not exist", regionID)
					sm.log.Error(errMsg)
					continue
				}
				r, exists, err := sm.regionMgr.GetRegionByID(regionID)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error locating region: %v", err.Error()))
					errMsg := fmt.Sprintf("kill region %v failed: %v", regionID, err.Error())
					sm.log.Error(errMsg)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, fmt.Sprintf("Region does not exist"))
					errMsg := fmt.Sprintf("kill region %v failed, region does not exist", regionID)
					sm.log.Error(errMsg)
					continue
				}

				h, err := sm.userMgr.RequestControlPermission(r, user)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error requesting permission: %v", err.Error()))
					errMsg := fmt.Sprintf("kill region %v failed: %v", regionID, err.Error())
					sm.log.Error(errMsg)
					continue
				}

				sm.hostMgr.KillRegionOnHost(r, h, func(success bool, message string) {
					if success {
						us.SignalSuccess(m.MessageID, message)
					} else {
						us.SignalError(m.MessageID, message)
					}
				})*/
			case "OpenConsole":
				/*regionID, err := m.ReadRegionID()
				if err != nil {
					us.SignalError(m.MessageID, "Invalid format")
					continue
				}
				sm.log.Info("User %v requesting region console %v", us.GetGUID(), regionID)
				user, exists, err := sm.userConn.GetUserByID(us.GetGUID())
				if err != nil {
					us.SignalError(m.MessageID, "Error looking up user")
					errMsg := fmt.Sprintf("region console %v failed, error finding requesting user", regionID)
					sm.log.Error(errMsg)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, "Invalid requesting user")
					errMsg := fmt.Sprintf("region console %v failed, requesting user does not exist", regionID)
					sm.log.Error(errMsg)
					continue
				}
				r, exists, err := sm.regionMgr.GetRegionByID(regionID)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error locating region: %v", err.Error()))
					errMsg := fmt.Sprintf("region console %v failed: %v", regionID, err.Error())
					sm.log.Error(errMsg)
					continue
				}
				if !exists {
					us.SignalError(m.MessageID, fmt.Sprintf("Region does not exist"))
					errMsg := fmt.Sprintf("region console %v failed, region does not exist", regionID)
					sm.log.Error(errMsg)
					continue
				}

				h, err := sm.userMgr.RequestControlPermission(r, user)
				if err != nil {
					us.SignalError(m.MessageID, fmt.Sprintf("Error requesting permission: %v", err.Error()))
					errMsg := fmt.Sprintf("region console %v failed: %v", regionID, err.Error())
					sm.log.Error(errMsg)
					continue
				}

				console = region.NewRestConsole(r, h)
				us.SignalSuccess(m.MessageID, "Console opened")*/
			case "CloseConsole":
				console.Close()
			case "DeleteJob":
				us.log.Info("Requesting delete job")
				id, err := m.ReadID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				var j mgm.Job
				exists := false
				for _, job := range us.mgm.GetJobs() {
					if job.ID == id {
						exists = true
						j = job
					}
				}
				if !exists {
					us.client.SignalError(m.MessageID, "Job does not exist")
					continue
				}
				if j.ID != id {
					us.client.SignalError(m.MessageID, "Job not found")
					continue
				}
				us.mgm.RemoveJob(j)
				//TODO some jobs may need files cleaned up... should we delete them here
				// or leave them and create a cleanup coroutine?
				us.client.SignalSuccess(m.MessageID, "Job Deleted")
			case "IarUpload":
				/*us.log.Info("Requesting iar upload")
				userID, password, err := m.ReadPassword()
				if err != nil {
					us.log.Error("Error reading iar upload request")
					continue
				}
				//isValid, err := sm.userConn.ValidatePassword(userID, password)
				//if err != nil {
				//	us.SignalError(m.MessageID, err.Error())
				//} else {
				//	if isValid {
				//password is valid, create the upload job
				users := us.mgm.GetUsers()
				exists := false
				var user mgm.User
				for _, u := range users {
					if u.UserID == userID {
						exists = true
						user = u
					}
				}
				if !exists {
					errMsg := fmt.Sprintf("Cannot creat job for load_iar: nonexistant user %v", userID)
					us.log.Error(errMsg)
					us.client.SignalError(m.MessageID, "User does not exist")
				}
				us.mgm.CreateLoadIarJob(user, "/")
				us.client.SignalSuccess(m.MessageID, "Job created")
				//	} else {
				//		us.SignalError(m.MessageID, "Invalid Password")
				//	}
				//}*/
			case "SetPassword":
				us.log.Info("Requesting password change")
				userID, password, err := m.ReadPassword()
				if err != nil {
					us.log.Error("Error reading password request")
					continue
				}
				if userID != us.client.GetGUID() && us.client.GetAccessLevel() < 250 {
					us.client.SignalError(m.MessageID, "Permission Denied")
					continue
				}
				if password == "" {
					us.client.SignalError(m.MessageID, "Password Cannot be blank")
					continue
				}
				var user mgm.User
				for _, u := range us.mgm.GetUsers() {
					if u.UserID == userID {
						user = u
					}
				}
				us.mgm.SetPassword(user, password)
				if err != nil {
					us.client.SignalError(m.MessageID, err.Error())
					continue
				}
				us.client.SignalSuccess(m.MessageID, "Password Set Successfully")
				us.log.Info("Password changed")

			case "GetDefaultConfig":
				/*us.log.Info("Requesting default configuration")
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
				}*/
			case "GetConfig":
				/*sm.log.Info("User %v requesting region configuration", us.GetGUID())
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
				}*/
			case "GetState":
				us.log.Info("Requesting state sync")
				uid := us.client.GetGUID()
				for _, u := range us.mgm.GetUsers() {
					us.client.Send(u)
				}
				for _, j := range us.mgm.GetJobs() {
					if j.User == uid {
						us.client.Send(j)
					}
				}
				if us.client.GetAccessLevel() > 249 {
					for _, pu := range us.mgm.GetPendingUsers() {
						us.client.Send(pu)
					}
				}
				estates := us.mgm.GetEstates()
				//calculate regions this user can control
				if us.client.GetAccessLevel() > 249 {
					//send them all
					for _, r := range us.mgm.GetRegions() {
						us.client.Send(r)
					}
					for _, e := range estates {
						us.client.Send(e)
					}
					for _, h := range us.mgm.GetHosts() {
						us.client.Send(h)
					}
				} else {
					//user must own or manage the estate in question
					mayManage := make(map[uuid.UUID]bool)
					for _, e := range estates {
						manage := false
						if e.Owner == uid {
							manage = true
						} else {
							for _, manager := range e.Managers {
								if manager == uid {
									manage = true
								}
							}
						}
						if manage == true {
							for _, rid := range e.Regions {
								mayManage[rid] = true
							}
						}
						//send estate to client
						us.client.Send(e)
					}
					for _, r := range us.mgm.GetRegions() {
						if _, ok := mayManage[r.UUID]; ok {
							us.client.Send(r)
						}
					}
				}

				for _, g := range us.mgm.GetGroups() {
					us.client.Send(g)
				}

				us.log.Info("State sync complete")
				//signal to the client that we have completed initial state sync
				us.client.SignalSuccess(m.MessageID, "State Sync Complete")

			default:
				us.log.Error("Error on message from client: ", m.MessageType)
				us.client.SignalError(m.MessageID, "Invalid request")
			}
		case <-us.client.GetClosingSignal():
			//the client connection has closed
			//host.Unsubscribe()
			//hostStats.Unsubscribe()
			us.closing <- us.client.GetGUID()
			return
		}

	}
}
