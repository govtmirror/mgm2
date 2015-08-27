package client

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
)

type registrant struct {
	Name     string
	Email    string
	Password string
	Template string
	Summary  string
}

func (r registrant) Validate() bool {
	if r.Name == "" {
		return false
	}
	names := strings.Split(r.Name, " ")
	if len(names) != 2 {
		return false
	}

	if r.Email == "" {
		return false
	}
	_, err := mail.ParseAddress(r.Email)
	if err != nil {
		return false
	}

	if r.Password == "" {
		return false
	}

	if r.Template != "M" && r.Template != "F" {
		return false
	}
	return true
}

// RegisterHandler is an http access point for user registration
func (m Manager) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid Request", http.StatusInternalServerError)
		return
	}

	var reg registrant
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !reg.Validate() {
		http.Error(w, "Invalid Request", http.StatusInternalServerError)
		return
	}

	/* Test if names are unique in pending users*/
	for _, user := range m.uMgr.GetPendingUsers() {
		if user.Email == reg.Email || user.Name == reg.Name {
			http.Error(w, "Credentials already exist", http.StatusInternalServerError)
			return
		}
	}

	/* Test if names are unique in registered users */
	for _, user := range m.uMgr.GetUsers() {
		if user.Email == reg.Email || user.Name == reg.Name {
			http.Error(w, "Credentials already exist", http.StatusInternalServerError)
			return
		}
	}

	m.uMgr.AddPendingUser(reg.Name, reg.Email, reg.Template, reg.Password, reg.Summary)

	//_ = hc.mailer.SendRegistrationSuccessful(reg.Name, reg.Email)
	//_ = hc.mailer.SendUserRegistered(reg.Name, reg.Email)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Success\": true}"))
}
