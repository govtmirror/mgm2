package webClient

import (
	"encoding/gob"
	"net/http"

	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/user"
	"github.com/m-o-s-e-s/mgm/email"
	"github.com/m-o-s-e-s/mgm/simian"
	"github.com/m-o-s-e-s/mgm/sql"
	"github.com/satori/go.uuid"
)

// HTTPConnector is a wrapper object for various http handlers
type HTTPConnector struct {
	authenticator simian.Connector
	logger        logger.Log
	userMgr       *user.Manager
	mailer        email.ClientEmailer
	jMgr          *job.Manager
	mgm           *sql.MGMDB
}

// NewHTTPConnector constructs an http connector for use
func NewHTTPConnector(jobMgr *job.Manager, mgm *sql.MGMDB, authenticator simian.Connector, userMgr *user.Manager, mailer email.ClientEmailer, log logger.Log) HTTPConnector {
	gob.Register(uuid.UUID{})

	return HTTPConnector{authenticator, logger.Wrap("HTTP", log), userMgr, mailer, jobMgr, mgm}
}

// PasswordResetHandler http endpoint for resseting a users password
func (hc HTTPConnector) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	/*
		type passwordReset struct {
			Name     string
			Token    uuid.UUID
			Password string
		}

		if r.Method != "POST" {
			http.Error(w, "Invalid Request", http.StatusInternalServerError)
			return
		}

		var req passwordReset
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if req.Password == "" {
			http.Error(w, "Invalid Request, Password cannot be null", http.StatusInternalServerError)
			return
		}

		user, exists, err := hc.authenticator.GetUserByName(req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "User does not exist", http.StatusInternalServerError)
			return
		}

		//test if account is still valid
		ids, err := hc.authenticator.GetIdentities(user.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, ident := range ids {
			if ident.Enabled == false {
				hc.logger.Error("Account is suspended")
				return
			}
		}

		isValid, err := hc.userMgr.ValidatePasswordToken(user.UserID, req.Token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !isValid {
			http.Error(w, "Invalid Request, bad token", http.StatusInternalServerError)
			return
		}

		err = hc.authenticator.SetPassword(user.UserID, req.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = hc.userMgr.ScrubPasswordToken(req.Token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = hc.mailer.SendPasswordResetEmail(user.Name, user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"Success\": true}"))
	*/
}

// PasswordTokenHandler is an http endpoing for requesting a password reset token
func (hc HTTPConnector) PasswordTokenHandler(w http.ResponseWriter, r *http.Request) {
	/*
		type tokenRequest struct {
			Email string
		}

		if r.Method != "POST" {
			http.Error(w, "Invalid Request", http.StatusInternalServerError)
			return
		}

		var req tokenRequest
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hc.logger.Info("Email reset requested for: %v from %v", req.Email, r.RemoteAddr)

		addr, err := mail.ParseAddress(req.Email)
		if err != nil {
			hc.logger.Error(err.Error())
			http.Error(w, "Invalid Request, formatting", http.StatusInternalServerError)
			return
		}

		user, exists, err := hc.authenticator.GetUserByEmail(addr.Address)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !exists {
			http.Error(w, "Invalid Request", http.StatusInternalServerError)
			return
		}

		ids, err := hc.authenticator.GetIdentities(user.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, ident := range ids {
			if ident.Enabled == false {
				hc.logger.Error("Account is suspended")
				return
			}
		}

		token, err := hc.userMgr.CreatePasswordResetToken(user.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		hc.logger.Info("Using %v for password reset token", token)
		err = hc.mailer.SendPasswordTokenEmail(user.Name, user.Email, token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"Success\": true}"))
	*/
}
