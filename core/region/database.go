package region

import (
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type regionDatabase struct {
	mysql database.Database
}

func (db regionDatabase) GetDefaultConfigs() ([]mgm.ConfigOption, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var cfgs []mgm.ConfigOption

	rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		c := mgm.ConfigOption{}
		err = rows.Scan(
			&c.Section,
			&c.Item,
			&c.Content,
		)
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}

func (db regionDatabase) GetConfigs(regionID uuid.UUID) ([]mgm.ConfigOption, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	var cfgs []mgm.ConfigOption

	rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region=?", regionID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		c := mgm.ConfigOption{}
		err = rows.Scan(
			&c.Section,
			&c.Item,
			&c.Content,
		)
		c.Region = regionID
		if err != nil {
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}
