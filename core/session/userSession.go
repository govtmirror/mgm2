package session

import (
	"fmt"
	"sync"

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
				go func() {
					if !isAdmin {
						us.client.SignalError(m.MessageID, "Permission Denied")
						return
					}
					address, err := m.ReadAddress()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
					}
					us.log.Info("Requesting add new Host %v", address)

					host := mgm.Host{}
					host.Address = address

					us.hMgr.AddHost(host, func(success bool, msg string) {
						if success {
							us.client.SignalSuccess(m.MessageID, msg)
						} else {
							us.client.SignalError(m.MessageID, msg)
						}
					})
				}()

			case "RemoveHost":
				go func() {
					if !isAdmin {
						us.client.SignalError(m.MessageID, "Permission Denied")
						return
					}
					hostID, err := m.ReadID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
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
						return
					}

					us.hMgr.RemoveHost(host, func(success bool, msg string) {
						if success {
							us.client.SignalSuccess(m.MessageID, msg)
						} else {
							us.client.SignalError(m.MessageID, msg)
						}
					})
				}()

			case "StartRegion":
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				us.log.Info("User %v requesting start region %v", uid, regionID)

				if !isAdmin {
					//check whitelist for control
					if !regionsWhitelist[regionID] {
						us.client.SignalError(m.MessageID, "Permission denied over region")
						continue
					}
				}
				//we can start the region
				var region mgm.Region
				found := false
				for _, r := range us.mgm.GetRegions() {
					if r.UUID == regionID {
						found = true
						region = r
					}
				}
				if !found {
					us.client.SignalError(m.MessageID, "Region not found")
					continue
				}

				var host mgm.Host
				found = false
				for _, h := range us.mgm.GetHosts() {
					if h.ID == region.Host {
						found = true
						host = h
					}
				}
				if !found {
					us.client.SignalError(m.MessageID, "Host not found")
					continue
				}

				if !host.Running {
					us.client.SignalError(m.MessageID, "Host is not running")
					continue
				}

				us.hMgr.StartRegionOnHost(region, host, func(success bool, msg string) {
					if success {
						us.client.SignalSuccess(m.MessageID, msg)
					} else {
						us.client.SignalError(m.MessageID, msg)
					}
				})

			case "StopRegion":
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				//locate region
				var r mgm.Region
				found := false
				for _, reg := range us.mgm.GetRegions() {
					if reg.UUID == regionID {
						found = true
						r = reg
					}
				}
				if !found {
					us.client.SignalError(m.MessageID, "Region does not exist")
					us.log.Info("User %v requesting stop region %v failed, region not found", uid, regionID)
					continue
				}
				us.log.Info("User %v requesting stop region %v", uid, regionID)
				if !isAdmin {
					//check if user has permission over this region
					if !regionsWhitelist[regionID] {
						us.client.SignalError(m.MessageID, "Permission Denied")
						us.log.Info("User %v requesting stop region %v failed, permission denied", uid, regionID)
						continue
					}
				}

				//lookup host record
				found = false
				var host mgm.Host
				for _, h := range us.mgm.GetHosts() {
					if h.ID == r.Host {
						found = true
						host = h
					}
				}
				if !found || r.Host == 0 {
					us.client.SignalError(m.MessageID, "Could not locate host, or region is not assigned to a host")
					us.log.Info("User %v requesting stop region %v failed, host not found", uid, regionID)
					continue
				}

				go func() {
					c, err := region.NewRestConsole(r, host)
					if err != nil {
						us.client.SignalError(m.MessageID, "Could not connect via console")
						us.log.Info("User %v requesting stop region %v failed, host not found", uid, regionID)
						return
					}
					c.Write("quit")
					c.Close()
					us.client.SignalSuccess(m.MessageID, "Region told to quit")
				}()

			case "KillRegion":
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				//locate region
				var region mgm.Region
				found := false
				for _, r := range us.mgm.GetRegions() {
					if r.UUID == regionID {
						found = true
						region = r
					}
				}
				if !found {
					us.client.SignalError(m.MessageID, "Region does not exist")
					us.log.Info("User %v requesting kill region %v failed, region not found", uid, regionID)
					continue
				}
				us.log.Info("User %v requesting kill region %v", uid, regionID)
				if !isAdmin {
					//check if user has permission over this region
					if !regionsWhitelist[regionID] {
						us.client.SignalError(m.MessageID, "Permission Denied")
						us.log.Info("User %v requesting kill region %v failed, permission denied", uid, regionID)
						continue
					}
				}

				//lookup host record
				found = false
				var host mgm.Host
				for _, h := range us.mgm.GetHosts() {
					if h.ID == region.Host {
						found = true
						host = h
					}
				}
				if !found || region.Host == 0 {
					us.client.SignalError(m.MessageID, "Could not locate host, or region is not assigned to a host")
					us.log.Info("User %v requesting kill region %v failed, host not found", uid, regionID)
					continue
				}

				us.hMgr.KillRegionOnHost(region, host, func(success bool, message string) {
					if success {
						us.client.SignalSuccess(m.MessageID, message)
					} else {
						us.client.SignalError(m.MessageID, message)
					}
				})
			case "OpenConsole":
				regionID, err := m.ReadRegionID()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid format")
					continue
				}
				us.log.Info("User %v requesting region console %v", uid, regionID)

				var r mgm.Region
				found := false
				for _, reg := range us.mgm.GetRegions() {
					if reg.UUID == regionID {
						found = true
						r = reg
					}
				}

				if !found {
					us.client.SignalError(m.MessageID, "Region does not exist")
					us.log.Info("User %v requesting kill region %v failed, region not found", uid, regionID)
					continue
				}
				if !isAdmin {
					if !regionsWhitelist[regionID] {
						us.client.SignalError(m.MessageID, "Permission Denied")
						us.log.Info("User %v requesting kill region %v failed, permission denied", uid, regionID)
						continue
					}
				}

				found = false
				var host mgm.Host
				for _, h := range us.mgm.GetHosts() {
					if h.ID == r.Host {
						found = true
						host = h
					}
				}

				if !found {
					us.client.SignalError(m.MessageID, "Host does not exist, or region is not assigned to a host")
					us.log.Info("User %v requesting kill region %v failed, region not found", uid, regionID)
					continue
				}

				if console != nil {
					us.log.Info("User %v closing existing console session", uid)
					console.Close()
				}

				console, err = region.NewRestConsole(r, host)

				go func() {
					for {
						select {
						case line, ok := <-console.Read():
							if !ok {
								return
							}
							us.client.Send(mgm.RegionConsole{regionID, line})
						}
					}
				}()

				us.log.Info("User %v open console complete, and in process", uid)

				if err != nil {
					us.client.SignalError(m.MessageID, fmt.Sprintf("Error opening console: %v", err.Error()))
				} else {
					us.client.SignalSuccess(m.MessageID, "Console opened")
				}

			case "ConsoleCommand":
				msg, err := m.ReadMessage()
				if err != nil {
					us.client.SignalError(m.MessageID, "Invalid message format")
					continue
				}
				us.log.Info("User %v sent console command %v", uid, msg)

				if console == nil {
					us.log.Info("User %v sending command with no active console", uid)
					us.client.SignalError(m.MessageID, "No active console")
					continue
				}

				console.Write(msg)

				us.client.SignalSuccess(m.MessageID, "Message forwarded to console")

			case "CloseConsole":
				us.log.Info("User %v requesting close console", uid)
				go func() {
					if console != nil {
						console.Close()
					}
				}()

			case "SetLocation":
				go func() {
					if !isAdmin {
						us.client.SignalError(m.MessageID, "Permission Denied")
						return
					}

					regionID, err := m.ReadRegionID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid id format")
						return
					}
					var reg mgm.Region
					found := false
					for _, r := range us.mgm.GetRegions() {
						if r.UUID == regionID {
							found = true
							reg = r
						}
					}
					if !found {
						us.client.SignalError(m.MessageID, "Region not found")
						return
					}

					x, y, err := m.ReadCoordinates()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid coordinate format")
						return
					}

					reg.LocX = x
					reg.LocY = y
					us.hMgr.UpdateRegion(reg, func(success bool, msg string) {
						if success {
							us.client.SignalSuccess(m.MessageID, msg)
						} else {
							us.client.SignalError(m.MessageID, msg)
						}
					})
				}()

			case "SetHost":
				go func() {
					if !isAdmin {
						us.client.SignalError(m.MessageID, "Permission Denied")
						return
					}
					//this can be assigning a region to a host, removing a region from a host, or both

					regionID, err := m.ReadRegionID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
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
						return
					}

					hostID, err := m.ReadID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
					}

					if hostID == region.Host {
						us.client.SignalError(m.MessageID, "Region is already on that host")
						return
					}

					us.log.Info("SetHost for region %v to host %v", region.UUID, hostID)

					abort := false
					abortMsg := ""
					var wg sync.WaitGroup

					//remove region from host if necessary
					us.log.Info("%v", region.Host)
					if region.Host != 0 {
						us.log.Info("Remove region from host should get called here")
						wg.Add(1)
						go func() {
							us.log.Info("Removing region %v from host %v", region.UUID, region.Host)
							for _, h := range us.mgm.GetHosts() {
								if h.ID == region.Host {
									us.hMgr.RemoveRegionFromHost(region, h, func(success bool, msg string) {
										if success {
											wg.Done()
											return
										}
										abort = true
										abortMsg = msg
										wg.Done()
										return
									})
								}
							}
						}()
						wg.Wait()
						if abort {
							us.client.SignalError(m.MessageID, abortMsg)
							return
						}
					}

					//assign region to new host if necessary
					if hostID != 0 {
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
							return
						}

						us.hMgr.AddRegionToHost(region, host, func(success bool, msg string) {
							if success {
								us.client.SignalSuccess(m.MessageID, msg)
							} else {
								us.client.SignalError(m.MessageID, msg)
							}
						})
					}

				}()

			case "SetEstate":
				go func() {
					if !isAdmin {
						us.client.SignalError(m.MessageID, "Permission Denied")
						return
					}

					estateID, err := m.ReadID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
					}
					regionID, err := m.ReadRegionID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
					}

					us.log.Info("Requesting add region %v to estate %v", regionID, estateID)

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
						return
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
						return
					}

					us.hMgr.SetRegionEstate(region, estate, func(success bool, msg string) {
						if success {
							us.client.SignalSuccess(m.MessageID, msg)
						} else {
							us.client.SignalError(m.MessageID, msg)
						}
					})

				}()

			case "DeleteJob":
				go func() {
					us.log.Info("Requesting delete job")
					id, err := m.ReadID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
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
						return
					}
					if j.ID != id {
						us.client.SignalError(m.MessageID, "Job not found")
						return
					}
					us.mgm.RemoveJob(j)
					//TODO some jobs may need files cleaned up... should we delete them here
					// or leave them and create a cleanup coroutine?
					us.client.SignalSuccess(m.MessageID, "Job Deleted")
				}()
			case "OarUpload":
				us.client.SignalError(m.MessageID, "Not Implemented")
			case "IarUpload":
				us.client.SignalError(m.MessageID, "Not Implemented")
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
				us.log.Info("Requesting default configuration")
				if isAdmin {
					cfgs := us.mgm.GetDefaultConfigs()
					for _, cfg := range cfgs {
						us.client.Send(cfg)
					}
					us.client.SignalSuccess(m.MessageID, "Default Config Retrieved")
					us.log.Info("User %v default configuration served", uid)
				} else {
					us.log.Info("User %v permission denied to default configurations", uid)
					us.client.SignalError(m.MessageID, "Permission Denied")
				}
			case "GetConfig":
				us.log.Info("Requesting region configuration")
				if isAdmin {
					regionID, err := m.ReadRegionID()
					if err != nil {
						us.client.SignalError(m.MessageID, "Invalid format")
						return
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
						return
					}

					cfgs := us.mgm.GetConfigs(region)
					for _, cfg := range cfgs {
						us.client.Send(cfg)
					}
					us.client.SignalSuccess(m.MessageID, "Region Config Retrieved")
					us.log.Info("User %v default configuration served", uid)
				} else {
					us.log.Info("User %v permission denied to default configurations", uid)
					us.client.SignalError(m.MessageID, "Permission Denied")
				}
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

					for _, s := range us.mgm.GetRegionStats() {
						us.client.Send(s)
					}

					for _, h := range us.mgm.GetHosts() {
						us.client.Send(h)
					}

					for _, s := range us.mgm.GetHostStats() {
						us.client.Send(s)
					}

				} else {
					//non admin, utilize whitelists
					for _, r := range us.mgm.GetRegions() {
						if regionsWhitelist[r.UUID] {
							us.client.Send(r)
						}
					}
					for _, r := range us.mgm.GetRegionStats() {
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
			us.closing <- uid
			return
		}

	}
}
