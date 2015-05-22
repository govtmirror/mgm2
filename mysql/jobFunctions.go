package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/M-O-S-E-S/mgm/mgm"
	//import mysql driver
	_ "github.com/go-sql-driver/mysql"
	"github.com/satori/go.uuid"
)

// GetJobByID retrieve a job record using the id of the job
func (db db) GetJobByID(id int) (mgm.Job, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return mgm.Job{}, err
	}
	defer con.Close()

	job := mgm.Job{}
	err = con.QueryRow("SELECT * FROM jobs WHERE id=?", id).Scan(&job.ID, &job.Timestamp, &job.Type, &job.User, &job.Data)
	if err != nil {
		return mgm.Job{}, err
	}

	return job, nil
}

// UpdateJob record an updated job record
func (db db) UpdateJob(job mgm.Job) error {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return err
	}
	defer con.Close()

	//The function states update job, but only the data field ever changes
	_, err = con.Exec("UPDATE jobs SET data=?", job.Data)
	if err != nil {
		return err
	}
	return nil
}

// DeleteJob purges a job record from the database
func (db db) DeleteJob(job mgm.Job) error {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return err
	}
	defer con.Close()

	_, err = con.Exec("DELETE FROM jobs WHERE id=?", job.ID)
	if err != nil {
		return err
	}
	return nil
}

// GetJobsForUser get all job records for a particular user
func (db db) GetJobsForUser(userID uuid.UUID) ([]mgm.Job, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query("SELECT * FROM jobs WHERE user=?", userID.String())
	defer rows.Close()
	if err != nil {
		db.log.Error("Error in database query: ", err.Error())
		return nil, err
	}

	var jobs []mgm.Job
	for rows.Next() {
		j := mgm.Job{}
		err = rows.Scan(
			&j.ID,
			&j.Timestamp,
			&j.Type,
			&j.User,
			&j.Data,
		)
		if err != nil {
			rows.Close()
			db.log.Error("Error in database query: ", err.Error())
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// CreateLoadIarJob utility function to create job of type load_iar
func (db db) CreateLoadIarJob(owner uuid.UUID, inventoryPath string) (mgm.Job, error) {
	loadIar := core.LoadIarJob{InventoryPath: "/"}
	data, err := json.Marshal(loadIar)
	if err != nil {
		return mgm.Job{}, err
	}
	return db.CreateJob("load_iar", owner, string(data))
}

// CreateJob create a new job record, returning the created job
func (db db) CreateJob(taskType string, userID uuid.UUID, data string) (mgm.Job, error) {
	con, err := sql.Open("mysql", fmt.Sprintf("%v:%v@tcp(%v:3306)/%v?parseTime=true", db.user, db.password, db.host, db.database))
	if err != nil {
		return mgm.Job{}, err
	}
	defer con.Close()

	res, err := con.Exec("INSERT INTO jobs (type, user, data) VALUES (?,?,?)", taskType, userID.String(), data)
	if err != nil {
		return mgm.Job{}, err
	}
	id, _ := res.LastInsertId()
	job := mgm.Job{}
	err = con.QueryRow("SELECT * FROM jobs WHERE id=?", id).Scan(&job.ID, &job.Timestamp, &job.Type, &job.User, &job.Data)
	if err != nil {
		return mgm.Job{}, err
	}

	return job, nil
}
