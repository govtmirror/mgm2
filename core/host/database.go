package host

import (
	"database/sql"
	"errors"

	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type hostDatabase struct {
	mysql persist.Database
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

	for i, h := range hosts {
		rows, err := con.Query("SELECT uuid FROM regions WHERE host=?", h.ID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			u := uuid.UUID{}
			err = rows.Scan(
				&u,
			)
			if err != nil {
				return nil, err
			}
			hosts[i].Regions = append(hosts[i].Regions, u)
		}
	}
	return hosts, nil
}

// GetHostByAddress retrieves a host record by address
func (db hostDatabase) GetHostByID(id int) (mgm.Host, bool, error) {
	h := mgm.Host{}
	if id == 0 {
		return h, false, errors.New("No assigned host")
	}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, false, err
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
		if err == sql.ErrNoRows {
			return h, false, nil
		}
		return h, false, err
	}

	rows, err := con.Query("SELECT uuid FROM regions WHERE host=?", h.ID)
	if err != nil {
		return h, false, err
	}
	defer rows.Close()
	for rows.Next() {
		u := uuid.UUID{}
		err = rows.Scan(
			&u,
		)
		if err != nil {
			return h, false, err
		}
		h.Regions = append(h.Regions, u)
	}

	return h, true, nil
}

// GetHostByAddress retrieves a host record by address
func (db hostDatabase) GetHostByAddress(address string) (mgm.Host, bool, error) {
	h := mgm.Host{}
	con, err := db.mysql.GetConnection()
	if err != nil {
		return h, false, err
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
		if err == sql.ErrNoRows {
			return h, false, nil
		}
		return h, false, err
	}

	rows, err := con.Query("SELECT uuid FROM regions WHERE host=?", h.ID)
	if err != nil {
		return h, false, err
	}
	defer rows.Close()
	for rows.Next() {
		u := uuid.UUID{}
		err = rows.Scan(
			&u,
		)
		if err != nil {
			return h, false, err
		}
		h.Regions = append(h.Regions, u)
	}

	return h, true, nil
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
