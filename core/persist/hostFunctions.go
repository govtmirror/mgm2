package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

func (m mgmDB) persistHost(host mgm.Host) {
	con, err := m.db.GetConnection()
	if err == nil {
		_, err = con.Exec("UPDATE hosts SET externalAddress=?, name=?, slots=? WHERE id=?",
			host.ExternalAddress, host.Hostname, host.Slots, host.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error persisting host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

func (m mgmDB) queryHosts() []mgm.Host {
	var hosts []mgm.Host
	con, err := m.db.GetConnection()
	if err != nil {
		errMsg := fmt.Sprintf("Error connecting to database: %v", err.Error())
		log.Fatal(errMsg)
		return hosts
	}
	defer con.Close()
	rows, err := con.Query("Select id, address, externalAddress, name, slots from hosts")
	if err != nil {
		errMsg := fmt.Sprintf("Error reading hosts: %v", err.Error())
		m.log.Error(errMsg)
		return hosts
	}
	defer rows.Close()
	for rows.Next() {
		h := mgm.Host{}
		err = rows.Scan(
			&h.ID,
			&h.Address,
			&h.ExternalAddress,
			&h.Hostname,
			&h.Slots,
		)
		if err != nil {
			errMsg := fmt.Sprintf("Error reading hosts: %v", err.Error())
			m.log.Error(errMsg)
			return hosts
		}
		hosts = append(hosts, h)
	}
	return hosts
}

func (m mgmDB) GetHosts() []mgm.Host {
	var hosts []mgm.Host
	r := mgmReq{}
	r.request = "GetHosts"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return hosts
		}
		hosts = append(hosts, h.(mgm.Host))
	}
}

func (m mgmDB) GetHostStats() []mgm.HostStat {
	var stats []mgm.HostStat
	r := mgmReq{}
	r.request = "GetHostStats"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return stats
		}
		stats = append(stats, h.(mgm.HostStat))
	}
}

func (m mgmDB) UpdateHost(host mgm.Host) {
	r := mgmReq{}
	r.request = "UpdateHost"
	r.object = host
	m.reqs <- r
}

func (m mgmDB) UpdateHostStat(stat mgm.HostStat) {
	r := mgmReq{}
	r.request = "UpdateHostStat"
	r.object = stat
	m.reqs <- r
}

func (m mgmDB) RemoveHost(host mgm.Host) {
	r := mgmReq{}
	r.request = "RemoveHost"
	r.object = host
	m.reqs <- r
}
