package client

import (
	"sync"

	"github.com/m-o-s-e-s/mgm/core/host"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/core/user"
	"github.com/satori/go.uuid"
)

// NewManager constructs a session manager for use
func NewManager(uMgr user.Manager, hMgr host.Manager, rMgr region.Manager, jMgr job.Manager, notify Notifier, log logger.Log) Manager {
	m := Manager{}
	m.log = logger.Wrap("CLIENT", log)
	m.uMgr = uMgr
	m.hMgr = hMgr
	m.rMgr = rMgr
	m.jMgr = jMgr

	m.clients = make(map[uuid.UUID]userConn)
	m.clientMutex = &sync.Mutex{}

	go listen(&m, notify)

	return m
}

// Manager is a central management point for client connections
type Manager struct {
	uMgr        user.Manager
	hMgr        host.Manager
	rMgr        region.Manager
	jMgr        job.Manager
	clients     map[uuid.UUID]userConn
	clientMutex *sync.Mutex
	log         logger.Log
}

func listen(m *Manager, n Notifier) {
	for {
		select {
		case id := <-n.hDel:
			m.HostRemoved(id)
		case h := <-n.hUp:
			m.HostAdded(h)
		case hs := <-n.hStat:
			m.HostStat(hs)
		}
	}
}
