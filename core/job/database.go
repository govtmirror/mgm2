package job

import (
	"encoding/json"

	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type jobDatabase struct {
	mysql database.Database
}

// GetJobByID retrieve a job record using the id of the job
func (db jobDatabase) GetJobByID(id int) (mgm.Job, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return mgm.Job{}, err
	}
	defer con.Close()

	j := mgm.Job{}
	err = con.QueryRow("SELECT * FROM jobs WHERE id=?", id).Scan(&j.ID, &j.Timestamp, &j.Type, &j.User, &j.Data)
	if err != nil {
		return mgm.Job{}, err
	}

	return j, nil
}

// UpdateJob record an updated job record
func (db jobDatabase) UpdateJob(job mgm.Job) error {
	con, err := db.mysql.GetConnection()
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
func (db jobDatabase) DeleteJob(job mgm.Job) error {
	con, err := db.mysql.GetConnection()
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
func (db jobDatabase) GetJobsForUser(userID uuid.UUID) ([]mgm.Job, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return nil, err
	}
	defer con.Close()

	rows, err := con.Query("SELECT * FROM jobs WHERE user=?", userID.String())
	defer rows.Close()
	if err != nil {
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
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, nil
}

// CreateLoadIarJob utility function to create job of type load_iar
func (db jobDatabase) CreateLoadIarJob(owner uuid.UUID, inventoryPath string) (mgm.Job, error) {
	loadIar := LoadIarJob{InventoryPath: "/"}
	data, err := json.Marshal(loadIar)
	if err != nil {
		return mgm.Job{}, err
	}
	return db.CreateJob("load_iar", owner, string(data))
}

// CreateJob create a new job record, returning the created job
func (db jobDatabase) CreateJob(taskType string, userID uuid.UUID, data string) (mgm.Job, error) {
	con, err := db.mysql.GetConnection()
	if err != nil {
		return mgm.Job{}, err
	}
	defer con.Close()

	res, err := con.Exec("INSERT INTO jobs (type, user, data) VALUES (?,?,?)", taskType, userID.String(), data)
	if err != nil {
		return mgm.Job{}, err
	}
	id, _ := res.LastInsertId()
	j := mgm.Job{}
	err = con.QueryRow("SELECT * FROM jobs WHERE id=?", id).Scan(&j.ID, &j.Timestamp, &j.Type, &j.User, &j.Data)
	if err != nil {
		return mgm.Job{}, err
	}

	return j, nil
}
