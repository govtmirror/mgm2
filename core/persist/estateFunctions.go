package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

func (m MGMDB) persistRegionEstate(region mgm.Region, estate mgm.Estate) {
	con, err := m.osdb.getConnection()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to database: %v", err.Error()))
		return
	}
	defer con.Close()

	_, err = con.Exec("REPLACE INTO estate_map VALUES (?,?)", region.UUID.String(), estate.ID)
	if err != nil {
		errMsg := fmt.Sprintf("Error updating estate_map: %v", err.Error())
		m.log.Error(errMsg)
	}
}

// QueryEstates retrieves all current estate records from mysql
func (m MGMDB) QueryEstates() []mgm.Estate {
	var estates []mgm.Estate
	con, err := m.osdb.getConnection()
	if err != nil {
		errMsg := fmt.Sprintf("Error connecting to database: %v", err.Error())
		log.Fatal(errMsg)
		return estates
	}
	defer con.Close()
	rows, err := con.Query("Select EstateID, EstateName, EstateOwner from estate_settings")
	if err != nil {
		errMsg := fmt.Sprintf("Error reading users: %v", err.Error())
		m.log.Error(errMsg)
		return estates
	}
	defer rows.Close()
	for rows.Next() {

		e := mgm.Estate{Managers: make([]uuid.UUID, 0), Regions: make([]uuid.UUID, 0)}
		err = rows.Scan(
			&e.ID,
			&e.Name,
			&e.Owner,
		)
		if err != nil {
			errMsg := fmt.Sprintf("Error reading estates: %v", err.Error())
			m.log.Error(errMsg)
			return estates
		}
		estates = append(estates, e)
	}

	//populate managers

	for i, e := range estates {
		//lookup managers
		rows, err := con.Query("SELECT uuid FROM estate_managers WHERE EstateID=?", e.ID)
		if err != nil {
			errMsg := fmt.Sprintf("Error reading estate managers: %v", err.Error())
			m.log.Error(errMsg)
			return estates
		}
		defer rows.Close()
		for rows.Next() {
			guid := uuid.UUID{}
			err = rows.Scan(&guid)
			if err != nil {
				errMsg := fmt.Sprintf("Error scanning estate managers: %v", err.Error())
				m.log.Error(errMsg)
				return estates
			}
			estates[i].Managers = append(estates[i].Managers, guid)
		}
		//lookup regions
		rows, err = con.Query("SELECT RegionID FROM estate_map WHERE EstateID=?", e.ID)
		defer rows.Close()
		if err != nil {
			errMsg := fmt.Sprintf("Error reading estate regions: %v", err.Error())
			m.log.Error(errMsg)
			return estates
		}
		for rows.Next() {
			guid := uuid.UUID{}
			err = rows.Scan(&guid)
			if err != nil {
				errMsg := fmt.Sprintf("Error scanning estate managers: %v", err.Error())
				m.log.Error(errMsg)
				return estates
			}
			estates[i].Regions = append(estates[i].Regions, guid)
		}
	}
	return estates
}
