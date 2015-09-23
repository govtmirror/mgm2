package sql

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

// hosts are created by clients inserting an ip address, that is all we can insert
func (m MGMDB) insertJob(job mgm.Job) (int64, error) {
	con, err := m.db.getConnection()
	var id int64
	if err != nil {
		return 0, err
	}
	defer con.Close()

	res, err := con.Exec("INSERT INTO jobs (type, user, data) VALUES (?,?,?)",
		job.Type, job.User.String(), job.Data)
	if err != nil {
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m MGMDB) persistJob(job mgm.Job) {
	con, err := m.db.getConnection()
	if err == nil {
		_, err = con.Exec("UPDATE jobs SET data=? WHERE id=?",
			job.Data, job.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error persisting host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

// PurgeJob remove a job from the database
func (m MGMDB) PurgeJob(job mgm.Job) {
	con, err := m.db.getConnection()
	if err == nil {
		_, err = con.Exec("DELETE FROM jobs WHERE id=?", job.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error purging host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

// QueryJobs reads all job records from the database
func (m MGMDB) QueryJobs() []mgm.Job {
	var jobs []mgm.Job
	con, err := m.db.getConnection()
	if err != nil {
		errMsg := fmt.Sprintf("Error connecting to database: %v", err.Error())
		log.Fatal(errMsg)
		return jobs
	}
	defer con.Close()
	rows, err := con.Query("Select * from jobs")
	if err != nil {
		errMsg := fmt.Sprintf("Error reading jobs: %v", err.Error())
		m.log.Error(errMsg)
		return jobs
	}
	defer rows.Close()
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
			errMsg := fmt.Sprintf("Error reading jobs: %v", err.Error())
			m.log.Error(errMsg)
			return jobs
		}
		jobs = append(jobs, j)
	}
	return jobs
}
