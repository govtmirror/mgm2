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

	rows, err := con.Query("Select address, port, name, slots, status from hosts")
	defer rows.Close()
	for rows.Next() {
		h := core.Host{}
		err = rows.Scan(
			&h.Address,
			&h.Port,
			&h.Hostname,
			&h.Slots,
			&h.Status,
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
