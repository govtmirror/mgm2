package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

// InsertHost creates a new host record by address, returning the row id
func (m MGMDB) InsertHost(address string) (int64, error) {
	con, err := m.db.getConnection()
	var id int64
	if err != nil {
		return 0, err
	}
	defer con.Close()

	res, err := con.Exec("INSERT INTO hosts (address) VALUES (?)",
		address)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m MGMDB) persistHost(host mgm.Host) {
	con, err := m.db.getConnection()
	if err == nil {
		_, err = con.Exec("UPDATE hosts SET externalAddress=?, name=?, slots=? WHERE id=?",
			host.ExternalAddress, host.Hostname, host.Slots, host.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error persisting host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

// PurgeHost removes a host record from the database
func (m MGMDB) PurgeHost(host int64) {
	con, err := m.db.getConnection()
	if err == nil {
		_, err = con.Exec("DELETE FROM hosts WHERE id=?", host)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error purging host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

// QueryHosts reads all host records from the database
func (m MGMDB) QueryHosts() []mgm.Host {
	var hosts []mgm.Host
	con, err := m.db.getConnection()
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
