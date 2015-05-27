package mysql

import (
	"database/sql"
	"fmt"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
	// load mysql driver
	_ "github.com/go-sql-driver/mysql"
)

//type RegionManager interface {
//	LoadedRegion(mgm.Region)
//}

type db struct {
	user     string
	password string
	database string
	host     string
	log      core.Logger
}

// Database is the database interface for persisting data
type Database interface {
	TestConnection() error

	GetDefaultConfigs() ([]mgm.ConfigOption, error)
	GetConfigs(regionID uuid.UUID) ([]mgm.ConfigOption, error)

	GetHosts() ([]mgm.Host, error)
	GetHostByAddress(address string) (mgm.Host, error)
	PlaceHostOffline(id uint) (mgm.Host, error)
	PlaceHostOnline(id uint) (mgm.Host, error)

	GetEstates() ([]mgm.Estate, error)

	GetRegions() ([]mgm.Region, error)
	GetRegionsOnHost(host mgm.Host) ([]mgm.Region, error)
	GetRegionsForUser(guid uuid.UUID) ([]mgm.Region, error)
	GetRegionByID(id uuid.UUID) (mgm.Region, error)

	GetJobsForUser(userID uuid.UUID) ([]mgm.Job, error)
	GetJobByID(id int) (mgm.Job, error)
	UpdateJob(job mgm.Job) error
	DeleteJob(job mgm.Job) error
	CreateJob(taskType string, userID uuid.UUID, data string) (mgm.Job, error)
	CreateLoadIarJob(owner uuid.UUID, inventoryPath string) (mgm.Job, error)

	GetPendingUsers() ([]mgm.PendingUser, error)
	AddPendingUser(name string, email string, template string, password string, summary string) error

	ValidatePasswordToken(userID uuid.UUID, token uuid.UUID) (bool, error)
	ScrubPasswordToken(token uuid.UUID) error
	ExpirePasswordTokens() error
	IsEmailUnique(email string) (bool, error)
	IsNameUnique(name string) (bool, error)
	CreatePasswordResetToken(userID uuid.UUID) (uuid.UUID, error)
}

// NewDatabase is a Database constructor
func NewDatabase(username string, password string, database string, host string, log core.Logger) Database {
	return db{username, password, database, host, log}
}

func (d db) TestConnection() error {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v", d.user, d.password, d.host, d.database))
	if err != nil {
		return err
	}
	defer con.Close()

	err = con.Ping()
	return err
}
