package client

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/googollee/go-socket.io"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type userResponse struct {
	Success bool
	Message string
}

type userConn struct {
	uid uuid.UUID
	sio socketio.Socket
	log logger.Log
}

func (c userConn) Close() {
	c.log.Info("Disconnected")
	c.sio.Emit("disconnect")
}

// NewClient register a new html5 websocket connection to the client manager
func (m Manager) NewClient(so socketio.Socket, guid uuid.UUID) {
	m.log.Info("New client: %v", guid.String())

	c := userConn{guid, so, logger.Wrap(guid.String(), m.log)}
	c.log.Info("Connected")

	m.clientMutex.Lock()
	if conn, ok := m.clients[guid]; ok {
		conn.Close()
	}
	m.clients[guid] = c
	m.clientMutex.Unlock()

	permissionDenied, _ := json.Marshal(userResponse{false, "Permission Denied"})
	success, _ := json.Marshal(userResponse{true, ""})

	so.On("AddHost", func(hostString string) string {
		c.log.Info("Requesting add host %v", hostString)
		// only admins may operate on hosts
		if !m.uMgr.UserIsAdmin(c.uid) {
			return string(permissionDenied)
		}
		err := m.hMgr.AddHost(hostString)
		if err != nil {
			resp, _ := json.Marshal(userResponse{false, err.Error()})
			return string(resp)
		}
		return string(success)
	})

	so.On("RemoveHost", func(idString string) string {
		c.log.Info("Requesting Remove host %v", idString)
		// only admins may operate on hosts
		if !m.uMgr.UserIsAdmin(c.uid) {
			return string(permissionDenied)
		}
		//parse host id from string
		id, err := strconv.ParseInt(idString, 10, 64)
		if err != nil {
			resp, _ := json.Marshal(userResponse{false, err.Error()})
			return string(resp)
		}
		//remove the host
		err = m.hMgr.RemoveHost(id)
		if err != nil {
			resp, _ := json.Marshal(userResponse{false, err.Error()})
			return string(resp)
		}
		return string(success)
	})

	so.On("StartRegion", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("StopRegion", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("KillRegion", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("OpenConsole", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("ConsoleCommand", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("CloseConsole", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("SetLocation", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("SetHost", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("SetEstate", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("DeleteJob", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("OarUpload", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("IarUpload", func(msg string) string {
		return string(permissionDenied)
	})

	so.On("SetPassword", func(msg string) string {
		type credentials struct {
			UserID   uuid.UUID
			Password string
		}
		creds := credentials{}
		err := json.Unmarshal([]byte(msg), &creds)
		if err != nil {
			resp, _ := json.Marshal(userResponse{false, "Invalid data packet"})
			return string(resp)
		}

		c.log.Info("Requesting change password for %v", creds.UserID.String())

		//only admins may change other users passwords
		if !m.uMgr.UserIsAdmin(c.uid) && creds.UserID != c.uid {
			return string(permissionDenied)
		}

		err = m.uMgr.SetPassword(creds.UserID, creds.Password)
		if err != nil {
			resp, _ := json.Marshal(userResponse{false, err.Error()})
			return string(resp)
		}
		return string(success)
	})

	so.On("GetConfig", func(guid string) string {
		c.log.Info("Requesting config %v", guid)
		if !m.uMgr.UserIsAdmin(c.uid) {
			return string(permissionDenied)
		}
		type response struct {
			Success bool
			Message string
			Configs []mgm.ConfigOption
		}
		state := response{}
		state.Configs = []mgm.ConfigOption{}
		if guid == "" {
			state.Configs = m.rMgr.GetDefaultConfigs()
			state.Success = true
		} else {
			id, err := uuid.FromString(guid)
			if err != nil {
				m.log.Error("Error serving region configs, invalid uuid")
				state.Message = fmt.Sprintf("Invalid Region ID %v", guid)
			} else {
				state.Configs = m.rMgr.GetConfigs(id)
				state.Success = true
			}
		}

		c.log.Info("Config complete with result %v", state.Success)
		result, _ := json.Marshal(state)
		return string(result)
	})

	so.On("GetState", func(msg string) string {
		c.log.Info("Requesting MGM State")

		type mgmState struct {
			Success      bool
			Users        []mgm.User
			PendingUsers []mgm.PendingUser
			Jobs         []mgm.Job
			Estates      []mgm.Estate
			Groups       []mgm.Group
			Regions      []mgm.Region
			RegionStats  []mgm.RegionStat
			Hosts        []mgm.Host
			HostStats    []mgm.HostStat
		}

		state := mgmState{}
		state.Success = true
		//populate data fields
		state.Users = m.uMgr.GetUsers()
		state.Jobs = m.jMgr.GetJobsForUser(c.uid)
		state.Estates = m.uMgr.GetEstates()
		state.Groups = m.uMgr.GetGroups()
		state.Regions = m.rMgr.GetRegions()
		state.RegionStats = m.rMgr.GetRegionStats()

		if m.uMgr.UserIsAdmin(c.uid) {
			state.PendingUsers = m.uMgr.GetPendingUsers()
			state.Hosts = m.hMgr.GetHosts()
			state.HostStats = m.hMgr.GetHostStats()
		}

		c.log.Info("Sending MGM state")
		result, _ := json.Marshal(state)
		return string(result)
	})
}
