package user

import (
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type simDatabase struct {
	mysql database.Database
}

// GetEstates retrieves all estates from mgm
func (db simDatabase) GetEstates() ([]mgm.Estate, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var estates []mgm.Estate

	rows, err := con.Query("Select EstateID, EstateName, EstateOwner from estate_settings")
	if err != nil {
		return nil, err
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
			return nil, err
		}
		estates = append(estates, e)
	}

	for i, e := range estates {
		//lookup managers
		rows, err := con.Query("SELECT uuid FROM estate_managers WHERE EstateID=?", e.ID)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			guid := uuid.UUID{}
			err = rows.Scan(&guid)
			if err != nil {
				return nil, err
			}
			estates[i].Managers = append(estates[i].Managers, guid)
		}
		//lookup regions
		rows, err = con.Query("SELECT RegionID FROM estate_map WHERE EstateID=?", e.ID)
		defer rows.Close()
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			guid := uuid.UUID{}
			err = rows.Scan(&guid)
			if err != nil {
				return nil, err
			}
			estates[i].Regions = append(estates[i].Regions, guid)
		}
	}

	return estates, nil
}
