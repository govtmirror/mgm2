package region

import (
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/mgm"
)

type simDatabase struct {
	mysql persist.Database
}

func (db simDatabase) GetConnectionString() string {
	return db.mysql.GetConnectionString()
}

// GetEstates retrieves all estates from mgm
func (db simDatabase) GetEstateNameForRegion(r mgm.Region) (string, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return "", err
	}
	defer con.Close()

	var e string

	err = con.QueryRow("Select EstateName from estate_settings, estate_map WHERE estate_settings.EstateID=estate_map.EstateID AND estate_map.RegionID=?", r.UUID.String()).Scan(&e)
	if err != nil {
		return e, err
	}
	return e, nil
}
