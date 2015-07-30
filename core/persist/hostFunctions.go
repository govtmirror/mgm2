package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

// hosts are created by clients inserting an ip address, that is all we can insert
func (m mgmDB) insertHost(host mgm.Host) (int64, error) {
	con, err := m.db.GetConnection()
	var id int64
	if err != nil {
		return 0, err
	}
	defer con.Close()

	res, err := con.Exec("INSERT INTO hosts (address) VALUES (?)",
		host.Address)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

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

func (m mgmDB) purgeHost(host mgm.Host) {
	con, err := m.db.GetConnection()
	if err == nil {
		_, err = con.Exec("DELETE FROM hosts WHERE id=?", host.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error purging host record: %v", err.Error())
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

func (m mgmDB) GetHost(id int64) (mgm.Host, bool) {
	for _, h := range m.GetHosts() {
		if h.ID == id {
			return h, true
		}
	}
	return mgm.Host{}, false
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

func (m mgmDB) GetHostStat(id int64) (mgm.HostStat, bool) {
	for _, st := range m.GetHostStats() {
		if st.ID == id {
			return st, true
		}
	}
	return mgm.HostStat{}, false
}

func (m mgmDB) UpdateHost(host mgm.Host) {
	r := mgmReq{}
	r.request = "UpdateHost"
	r.object = host
	m.reqs <- r
}

func (m mgmDB) AddHost(host mgm.Host) mgm.Host {
	r := mgmReq{}
	r.request = "AddHost"
	r.result = make(chan interface{}, 2)
	r.object = host
	m.reqs <- r
	resp := <-r.result
	return resp.(mgm.Host)
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
