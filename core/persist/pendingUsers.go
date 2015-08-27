package persist

import (
	"fmt"
	"log"

	"github.com/m-o-s-e-s/mgm/mgm"
)

// QueryPendingUsers reads pending user records from the database
func (m MGMDB) QueryPendingUsers() []mgm.PendingUser {
	var users []mgm.PendingUser
	con, err := m.db.getConnection()
	if err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to database: %v", err.Error()))
		return users
	}
	defer con.Close()
	rows, err := con.Query("Select * from users")
	if err != nil {
		log.Fatal(fmt.Sprintf("Error getting pending users: %v", err.Error()))
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
			log.Fatal(fmt.Sprintf("Error parsing pending users: %v", err.Error()))
			return users
		}
		users = append(users, u)
	}
	return users
}
