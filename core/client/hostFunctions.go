package client

import "github.com/m-o-s-e-s/mgm/mgm"

// HostAdded notifies connected clients that a host has been added
func (m Manager) HostAdded(h mgm.Host) {
	m.clientMutex.Lock()
	defer m.clientMutex.Unlock()

	for _, c := range m.clients {
		go func(conn userConn, host mgm.Host) {
			if m.uMgr.UserIsAdmin(conn.uid) {
				conn.sio.Emit("Host", host)
			}
		}(c, h)
	}
}

// HostRemoved notifies connected clients that a host has been removed
func (m Manager) HostRemoved(id int64) {
	m.clientMutex.Lock()
	defer m.clientMutex.Unlock()

	for _, c := range m.clients {
		go func(conn userConn, hid int64) {
			if m.uMgr.UserIsAdmin(conn.uid) {
				conn.sio.Emit("HostRemoved", hid)
			}
		}(c, id)
	}
}

// HostStat notifies connected clients that a host has been removed
func (m Manager) HostStat(hs mgm.HostStat) {
	m.clientMutex.Lock()
	defer m.clientMutex.Unlock()

	for _, c := range m.clients {
		go func(conn userConn, stat mgm.HostStat) {
			if m.uMgr.UserIsAdmin(conn.uid) {
				conn.sio.Emit("HostStat", stat)
			}
		}(c, hs)
	}
}
