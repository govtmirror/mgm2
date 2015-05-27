package mysql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/M-O-S-E-S/mgm/mgm"
	"github.com/satori/go.uuid"
)

// GetHosts retrieves all host records from the database
func (db db) GetHosts() ([]mgm.Host, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
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
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		regions, _ := db.GetRegionsOnHost(h)
		var regids []uuid.UUID
		for _, r := range regions {
			regids = append(regids, r.UUID)
		}
		h.Regions = regids
		hosts = append(hosts, h)
	}
	return hosts, nil
}

// GetHostByAddress retrieves a host record by address
func (db db) GetHostByAddress(address string) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
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
		db.log.Error("Error in database query: ", err.Error())
		return h, err
	}
	return h, nil
}

// PlaceHostOffline sets the specified host to offline, and returns the updated struct
func (db db) PlaceHostOffline(id uint) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return h, err
	}
	defer con.Close()

	_, err = con.Exec("UPDATE hosts SET running=? WHERE id=?", false, id)
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
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
		db.log.Error("Error in database query: ", err.Error())
		return h, err
	}
	return h, nil
}

// PlaceHostOnline sets the specified host to online, and returns the updated struct
func (db db) PlaceHostOnline(id uint) (mgm.Host, error) {
	h := mgm.Host{}
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return h, err
	}
	defer con.Close()

	_, err = con.Exec("UPDATE hosts SET running=? WHERE id=?", true, id)
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
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
		db.log.Error("Error in database query: ", err.Error())
		return h, err
	}
	return h, nil
}
