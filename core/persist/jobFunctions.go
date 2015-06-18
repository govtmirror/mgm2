package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

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
