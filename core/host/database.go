package host

import (
	"errors"

	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
)

type hostDatabase struct {
	mysql database.Database
}

// GetHosts retrieves all host records from the database
func (db hostDatabase) GetHosts() ([]mgm.Host, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var hosts []mgm.Host

	rows, err := con.Query("Select id, address, port, name, slots, running from hosts")
	defer rows.Close()
	for rows.Next() {
		h := mgm.Host{}
		err = rows.Scan(
			&h.ID,
			&h.Address,
			&h.Port,
			&h.Hostname,
			&h.Slots,
			&h.Running,
		)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// GetHostByAddress retrieves a host record by address
func (db hostDatabase) GetHostByAddress(address string) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, err
	}
	defer con.Close()

	err = con.QueryRow("SELECT id, address, port, name, slots, running FROM hosts WHERE address=?", address).Scan(
		&h.ID,
		&h.Address,
		&h.Port,
		&h.Hostname,
		&h.Slots,
		&h.Running,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return h, errors.New("Host not found")
		}
		return h, err
	}
	return h, nil
}

// PlaceHostOffline sets the specified host to offline, and returns the updated struct
func (db hostDatabase) PlaceHostOffline(id uint) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, err
	}
	defer con.Close()

	_, err = con.Exec("UPDATE hosts SET running=? WHERE id=?", false, id)
	if err != nil {
		return h, err
	}

	err = con.QueryRow("SELECT id, address, port, name, slots, running FROM hosts WHERE id=?", id).Scan(
		&h.ID,
		&h.Address,
		&h.Port,
		&h.Hostname,
		&h.Slots,
		&h.Running,
	)
	if err != nil {
		return h, err
	}
	return h, nil
}

// PlaceHostOnline sets the specified host to online, and returns the updated struct
func (db hostDatabase) PlaceHostOnline(id uint) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, err
	}
	defer con.Close()

	_, err = con.Exec("UPDATE hosts SET running=? WHERE id=?", true, id)
	if err != nil {
		return h, err
	}

	err = con.QueryRow("SELECT id, address, port, name, slots, running FROM hosts WHERE id=?", id).Scan(
		&h.ID,
		&h.Address,
		&h.Port,
		&h.Hostname,
		&h.Slots,
		&h.Running,
	)
	if err != nil {
		return h, err
	}
	return h, nil
}

// GetRegionsOnHost retrieves all region records for a specified host
func (db hostDatabase) GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query(
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, externalAddress, slaveAddress, isRunning from regions "+
			"where slaveAddress=?", host.Address)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var regions []mgm.Region
	for rows.Next() {
		r := mgm.Region{}
		err = rows.Scan(
			&r.UUID,
			&r.Name,
			&r.Size,
			&r.HTTPPort,
			&r.ConsolePort,
			&r.ConsoleUname,
			&r.ConsolePass,
			&r.LocX,
			&r.LocY,
			&r.ExternalAddress,
			&r.SlaveAddress,
			&r.IsRunning,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}
