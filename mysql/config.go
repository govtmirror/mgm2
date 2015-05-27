package mysql

import (
	"database/sql"
	"fmt"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

func (db db) GetDefaultConfigs() ([]mgm.ConfigOption, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
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
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}

func (db db) GetConfigs(regionID uuid.UUID) ([]mgm.ConfigOption, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", db.user, db.password, db.host, db.database))
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
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		cfgs = append(cfgs, c)
	}
	return cfgs, nil
}
