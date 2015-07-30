package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

// hosts are created by clients inserting an ip address, that is all we can insert
func (m mgmDB) insertJob(job mgm.Job) (int64, error) {
	con, err := m.db.GetConnection()
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

func (m mgmDB) persistJob(job mgm.Job) {
	con, err := m.db.GetConnection()
	if err == nil {
		_, err = con.Exec("UPDATE jobs SET data=? WHERE id=?",
			job.Data, job.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error persisting host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

func (m mgmDB) purgeJob(job mgm.Job) {
	con, err := m.db.GetConnection()
	if err == nil {
		_, err = con.Exec("DELETE FROM jobs WHERE id=?", job.ID)
	}
	if err != nil {
		errMsg := fmt.Sprintf("Error purging host record: %v", err.Error())
		m.log.Error(errMsg)
	}
}

func (m mgmDB) queryJobs() []mgm.Job {
	var jobs []mgm.Job
	con, err := m.db.GetConnection()
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

func (m mgmDB) GetJobs() []mgm.Job {
	var jobs []mgm.Job
	r := mgmReq{}
	r.request = "GetJobs"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return jobs
		}
		jobs = append(jobs, h.(mgm.Job))
	}
}

func (m mgmDB) AddJob(j mgm.Job) int64 {
	r := mgmReq{}
	r.request = "AddJob"
	r.object = j
	r.result = make(chan interface{}, 2)
	m.reqs <- r
	obj := <-r.result
	job := obj.(mgm.Job)
	return job.ID
}

func (m mgmDB) UpdateJob(j mgm.Job) {
	r := mgmReq{}
	r.request = "UpdateJob"
	r.object = j
	m.reqs <- r
}

func (m mgmDB) RemoveJob(j mgm.Job) {
	r := mgmReq{}
	r.request = "RemoveJob"
	r.object = j
	m.reqs <- r
}
