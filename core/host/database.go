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

	rows, err := con.Query("Select id, address, externalAddress, name, slots from hosts")
	if err != nil {
		return nil, err
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
			return nil, err
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// GetHostByAddress retrieves a host record by address
func (db hostDatabase) GetHostByID(id uint) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, err
	}
	defer con.Close()

	err = con.QueryRow("SELECT id, address, externalAddress, name, slots FROM hosts WHERE id=?", id).Scan(
		&h.ID,
		&h.Address,
		&h.ExternalAddress,
		&h.Hostname,
		&h.Slots,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return h, errors.New("Host not found")
		}
		return h, err
	}
	return h, nil
}

// GetHostByAddress retrieves a host record by address
func (db hostDatabase) GetHostByAddress(address string) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, err
	}
	defer con.Close()

	err = con.QueryRow("SELECT id, address, externalAddress, name, slots FROM hosts WHERE address=?", address).Scan(
		&h.ID,
		&h.Address,
		&h.ExternalAddress,
		&h.Hostname,
		&h.Slots,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return h, errors.New("Host not found")
		}
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
		"Select uuid, name, size, httpPort, consolePort, consoleUname, consolePass, locX, locY, host from regions "+
			"where host=?", host.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
			&r.Host,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}
		regions = append(regions, r)
	}
	return regions, nil
}

func (db hostDatabase) UpdateHost(h mgm.Host, reg Registration) (mgm.Host, error) {
	con, err := db.mysql.GetConnection()

	_, err = con.Exec("UPDATE hosts SET externalAddress=?, name=?, slots=? WHERE id=?",
		reg.ExternalAddress, reg.Name, reg.Slots, h.ID)
	if err != nil {
		return h, err
	}
	h.ExternalAddress = reg.ExternalAddress
	h.Hostname = reg.Name
	h.Slots = reg.Slots
	return h, nil
}
