package persist

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

func (m mgmDB) GetUsers() []mgm.User {
	var users []mgm.User
	r := mgmReq{}
	r.request = "GetUsers"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return users
		}
		users = append(users, h.(mgm.User))
	}
}

func (m mgmDB) queryPendingUsers() []mgm.PendingUser {
	var users []mgm.PendingUser
	con, err := m.db.GetConnection()
	if err != nil {
		errMsg := fmt.Sprintf("Error connecting to database: %v", err.Error())
		log.Fatal(errMsg)
		return users
	}
	defer con.Close()
	rows, err := con.Query("Select * from users")
	if err != nil {
		errMsg := fmt.Sprintf("Error reading users: %v", err.Error())
		m.log.Error(errMsg)
		return users
	}
	defer rows.Close()
	for rows.Next() {
		u := mgm.PendingUser{}
		err = rows.Scan(
			&u.Name,
			&u.Email,
			&u.Gender,
			&u.PasswordHash,
			&u.Summary,
		)
		if err != nil {
			errMsg := fmt.Sprintf("Error reading hosts: %v", err.Error())
			m.log.Error(errMsg)
			return users
		}
		users = append(users, u)
	}
	return users
}

func (m mgmDB) GetPendingUsers() []mgm.PendingUser {
	var users []mgm.PendingUser
	r := mgmReq{}
	r.request = "GetPendingUsers"
	r.result = make(chan interface{}, 64)
	m.reqs <- r
	for {
		h, ok := <-r.result
		if !ok {
			return users
		}
		users = append(users, h.(mgm.PendingUser))
	}
}

func (m mgmDB) UpdateUser(user mgm.User) {
	r := mgmReq{}
	r.request = "UpdateUser"
	r.object = user
	m.reqs <- r
}

func (m mgmDB) SetPassword(user mgm.User, password string) {
	hasher := md5.New()
	hasher.Write([]byte(password))

	r := mgmReq{}
	r.request = "UpdateCredential"
	r.object = mgm.UserCredential{
		UserID: user.UserID,
		Hash:   "$1$" + hex.EncodeToString(hasher.Sum(nil)),
	}
	m.reqs <- r
}
