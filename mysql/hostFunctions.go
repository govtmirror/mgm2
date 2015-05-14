package mysql

import (
	"database/sql"
	"fmt"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/satori/go.uuid"
)

// GetHosts retrieves all host records from the database
func (db Database) GetHosts() ([]core.Host, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var hosts []core.Host

	rows, err := con.Query("Select id, address, port, name, slots, running from hosts")
	defer rows.Close()
	for rows.Next() {
		h := core.Host{}
		err = rows.Scan(
			&h.ID,
			&h.Address,
			&h.Port,
			&h.Hostname,
			&h.Slots,
			&h.Running,
		)
		if err != nil {
			fmt.Println(err)
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
func (db Database) GetHostByAddress(address string) (core.Host, error) {
	h := core.Host{}
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
		fmt.Println(err)
		return h, err
	}
	return h, nil
}

// PlaceHostOffline sets the specified host to offline, and returns the updated struct
func (db Database) PlaceHostOffline(id uint) (core.Host, error) {
	h := core.Host{}
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return h, err
	}
	defer con.Close()

	_, err = con.Exec("UPDATE hosts SET running=? WHERE id=?", false, id)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return h, err
	}
	return h, nil
}

// PlaceHostOnline sets the specified host to online, and returns the updated struct
func (db Database) PlaceHostOnline(id uint) (core.Host, error) {
	h := core.Host{}
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
	if err != nil {
		return h, err
	}
	defer con.Close()

	_, err = con.Exec("UPDATE hosts SET running=? WHERE id=?", true, id)
	if err != nil {
		fmt.Println(err)
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
		fmt.Println(err)
		return h, err
	}
	return h, nil
}
