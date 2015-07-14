package session

import (
	"fmt"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type userSession struct {
	client   core.UserSession
	closing  chan<- uuid.UUID
	mgm      persist.MGMDB
	log      logger.Log
	hMgr     host.Manager
	notifier Notifier
}

func (us userSession) process() {

	clientMsg := make(chan []byte, 32)

	go us.client.Read(clientMsg)

	var console region.RestConsole

	isAdmin := us.client.GetAccessLevel() > 249
	uid := us.client.GetGUID()

	// if we arent admin, maintain a list of estates and regions that we manage
	regionsWhitelist := make(map[uuid.UUID]bool)
	estatesWhitelist := make(map[int64]bool)

	if !isAdmin {
		//populate the whitelists
		for _, e := range us.mgm.GetEstates() {
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
				estatesWhitelist[e.ID] = true
				for _, id := range e.Regions {
					regionsWhitelist[id] = true
				}
			} else {
				estatesWhitelist[e.ID] = false
				for _, id := range e.Regions {
					regionsWhitelist[id] = false
				}
			}
		}
	}

	for {
		select {
		//MGM EVENTS
		case h := <-us.notifier.hDel:
			if isAdmin {
				us.client.Send(mgm.HostDeleted{h.ID})
			}
		case h := <-us.notifier.hUp:
			if isAdmin {
				us.client.Send(h)
			}
		case s := <-us.notifier.hStat:
			if isAdmin {
				us.client.Send(s)
			}
		case r := <-us.notifier.rUp:
			// new or updated region
			if regionsWhitelist[r.UUID] || isAdmin {
				us.client.Send(r)
			}
		case r := <-us.notifier.rDel:
			if regionsWhitelist[r.UUID] || isAdmin {
				us.client.Send(mgm.RegionDeleted{r.UUID})
			}
		case s := <-us.notifier.rStat:
			if regionsWhitelist[s.UUID] || isAdmin {
				us.client.Send(s)
			}
		case e := <-us.notifier.eUp:
			us.log.Info("Sending estate update to client")
			//make sure we still manage it
			if isAdmin {
				us.client.Send(e)
			} else {
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
				if manage {
					us.client.Send(e)
				}
				estatesWhitelist[e.ID] = manage
			}
		case e := <-us.notifier.eDel:
			us.client.Send(mgm.EstateDeleted{e.ID})
			estatesWhitelist[e.ID] = false

		// COMMANDS FROM THE CLIENT
		case msg := <-clientMsg:
			//message from client
			m := mgm.UserMessage{}
			m.Load(msg)
			switch m.MessageType {
			case "AddHost":
				if !isAdmin {
					us.client.SignalError(m.MessageID, "Permission Denied")
					continue
				}
				address, err := m.ReadAddress()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				us.log.Info("Requesting add new Host %v", address)
				//double-check for duplicates:
				for _, h := range us.mgm.GetHosts() {
					if h.Address == address {
						us.client.SignalError(m.MessageID, "Host already exists")
						continue
					}
				}
				host := mgm.Host{}
				host.Address = address
				us.mgm.AddHost(host)
				us.client.SignalSuccess(m.MessageID, "Host added")
			case "RemoveHost":
				if !isAdmin {
					us.client.SignalError(m.MessageID, "Permission Denied")
					continue
				}
				hostID, err := m.ReadID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				us.log.Info("Requesting remove Host %v", hostID)
				var host mgm.Host
				exists := false
				for _, h := range us.mgm.GetHosts() {
					if h.ID == hostID {
						host = h
						exists = true
					}
				}
				if !exists {
					us.client.SignalError(m.MessageID, "Host does not exist")
					errMsg := fmt.Sprintf("delete host %v failed, host does not exist", hostID)
					us.log.Error(errMsg)
					continue
				}

				us.mgm.RemoveHost(host)
				us.client.SignalSuccess(m.MessageID, "host removed")

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

			case "SetHost":
				if !isAdmin {
					us.client.SignalError(m.MessageID, "Permission Denied")
					continue
				}

				regionID, err := m.ReadRegionID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				var region mgm.Region
				found := false
				for _, r := range us.mgm.GetRegions() {
					if r.UUID == regionID {
						region = r
						found = true
					}
				}
				if !found {
					us.client.SignalError(m.MessageID, "Region not found")
					continue
				}

				hostID, err := m.ReadID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				var host mgm.Host
				found = false
				for _, h := range us.mgm.GetHosts() {
					if h.ID == hostID {
						host = h
						found = true
					}
				}
				if !found && hostID != 0 {
					us.client.SignalError(m.MessageID, "Host not found")
					continue
				}

				us.mgm.MoveRegionToHost(region, host)

				us.client.SignalError(m.MessageID, "region flagged for new host")

			case "SetEstate":
				if !isAdmin {
					us.client.SignalError(m.MessageID, "Permission Denied")
					continue
				}

				estateID, err := m.ReadID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}

				us.log.Info("Requesting add region %v to estate %v", regionID, estateID)

				//Must be admin to mod region estates
				if !isAdmin {
					us.client.SignalError(m.MessageID, "Permission Denied")
					us.log.Error("Add region to estate failed, permission denied")
					continue
				}

				var region mgm.Region
				regionFound := false
				var estate mgm.Estate
				estateFound := false
				for _, r := range us.mgm.GetRegions() {
					if r.UUID == regionID {
						regionFound = true
						region = r
					}
				}
				if !regionFound {
					us.client.SignalError(m.MessageID, "Region does not exist")
					us.log.Error("Add region to estate failed, region not found")
					continue
				}
				for _, e := range us.mgm.GetEstates() {
					if e.ID == estateID {
						estateFound = true
						estate = e
					}
				}
				if !estateFound {
					us.client.SignalError(m.MessageID, "Estate does not exist")
					us.log.Error("Add region to estate failed, estate not found")
					continue
				}

				go us.mgm.MoveRegionToEstate(region, estate)
				us.client.SignalSuccess(m.MessageID, "Region Flagged for new estate")

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
				if userID != uid && !isAdmin {
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
				for _, u := range us.mgm.GetUsers() {
					us.client.Send(u)
				}
				for _, j := range us.mgm.GetJobs() {
					if j.User == uid {
						us.client.Send(j)
					}
				}
				for _, e := range us.mgm.GetEstates() {
					us.client.Send(e)
				}
				for _, g := range us.mgm.GetGroups() {
					us.client.Send(g)
				}

				if isAdmin {
					for _, pu := range us.mgm.GetPendingUsers() {
						us.client.Send(pu)
					}

					for _, r := range us.mgm.GetRegions() {
						us.client.Send(r)
					}

					for _, h := range us.mgm.GetHosts() {
						us.client.Send(h)
					}

				} else {
					//non admin, utilize whitelists
					for _, r := range us.mgm.GetRegions() {
						if regionsWhitelist[r.UUID] {
							us.client.Send(r)
						}
					}
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
			us.closing <- us.client.GetGUID()
			return
		}

	}
}
