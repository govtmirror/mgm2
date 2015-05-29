package webClient

import (
	"encoding/gob"
	"encoding/json"
	"net/http"
	"net/mail"

	"github.com/gorilla/sessions"
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/job"
	"github.com/m-o-s-e-s/mgm/core/user"
	"github.com/m-o-s-e-s/mgm/email"
	"github.com/m-o-s-e-s/mgm/simian"
	"github.com/satori/go.uuid"
)

type clientAuthResponse struct {
	UUID        uuid.UUID
	AccessLevel uint8
	Message     string
	Success     bool
}

// HTTPConnector responds to http client requests
type HTTPConnector interface {
	PasswordTokenHandler(w http.ResponseWriter, r *http.Request)
	PasswordResetHandler(w http.ResponseWriter, r *http.Request)
	RegisterHandler(w http.ResponseWriter, r *http.Request)
	ResumeHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
	LoginHandler(w http.ResponseWriter, r *http.Request)
	UploadHandler(w http.ResponseWriter, r *http.Request)
	GetStore() *sessions.CookieStore
}

type httpConn struct {
	store         *sessions.CookieStore
	authenticator simian.Connector
	logger        core.Logger
	userMgr       user.Manager
	mailer        email.ClientEmailer
	jMgr          job.Manager
}

// NewHTTPConnector constructs an http connector for use
func NewHTTPConnector(sessionKey string, jobMgr job.Manager, authenticator simian.Connector, userMgr user.Manager, mailer email.ClientEmailer, logger core.Logger) HTTPConnector {
	gob.Register(uuid.UUID{})

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:   "/",
		MaxAge: 3600 * 8,
	}
	return httpConn{store, authenticator, logger, userMgr, mailer, jobMgr}
}

func (hc httpConn) GetStore() *sessions.CookieStore {
	return hc.store
}

func (hc httpConn) LoginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	type clientAuthRequest struct {
		Username string
		Password string
	}

	var t clientAuthRequest
	err := decoder.Decode(&t)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	valid, guid, err := hc.authenticator.Auth(t.Username, t.Password)
	if err != nil {
		if valid == false {
			http.Error(w, "Invalid Credential", http.StatusInternalServerError)
			return
		}
		response := clientAuthResponse{Message: err.Error()}
		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	user, exists, err := hc.authenticator.GetUserByID(guid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !exists {
		http.Error(w, "Invalid Credentials", http.StatusInternalServerError)
		return
	}

	s, _ := hc.store.Get(r, "MGM")
	s.Values["guid"] = guid
	s.Values["address"] = r.RemoteAddr
	s.Values["ulevel"] = user.AccessLevel
	err = s.Save(r, w)
	if err != nil {
		hc.logger.Error("Error in httpConnector: ", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hc.logger.Info("Session saved, returning success")

	response := clientAuthResponse{guid, user.AccessLevel, "", true}
	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (hc httpConn) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := hc.store.Get(r, "MGM")
	delete(s.Values, "guid")
	delete(s.Values, "address")
	s.Save(r, w)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Success\": true}"))
}

func (hc httpConn) ResumeHandler(w http.ResponseWriter, r *http.Request) {
	s, _ := hc.store.Get(r, "MGM")

	if s.Values["guid"] == nil {
		response := clientAuthResponse{}
		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	response := clientAuthResponse{s.Values["guid"].(uuid.UUID), s.Values["ulevel"].(uint8), "", true}
	js, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}

func (hc httpConn) RegisterHandler(w http.ResponseWriter, r *http.Request) {

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
	unique, err := hc.userMgr.IsEmailUnique(reg.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !unique {
		http.Error(w, "Credentials already exist", http.StatusInternalServerError)
		return
	}

	unique, err = hc.userMgr.IsNameUnique(reg.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !unique {
		http.Error(w, "Credentials already exist", http.StatusInternalServerError)
		return
	}

	/* Test if names are unique in registered users */
	exists, err := hc.authenticator.IsEmailTaken(reg.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Email already exists", http.StatusInternalServerError)
		return
	}
	exists, err = hc.authenticator.IsNameTaken(reg.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Name already exists", http.StatusInternalServerError)
		return
	}

	err = hc.userMgr.AddPendingUser(reg.Name, reg.Email, reg.Template, reg.Password, reg.Summary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = hc.mailer.SendRegistrationSuccessful(reg.Name, reg.Email)
	_ = hc.mailer.SendUserRegistered(reg.Name, reg.Email)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Success\": true}"))
}

func (hc httpConn) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
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
}

func (hc httpConn) PasswordTokenHandler(w http.ResponseWriter, r *http.Request) {
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
}
