package webClient

import (
  "net/http"
  "net/mail"
  "encoding/json"
  "github.com/satori/go.uuid"
  "github.com/gorilla/sessions"
  "github.com/M-O-S-E-S/mgm2/core"
  "encoding/gob"
)

type Logger interface {
  Trace(format string, v ...interface{})
  Debug(format string, v ...interface{})
  Info(format string, v ...interface{})
  Warn(format string, v ...interface{})
  Error(format string, v ...interface{})
  Fatal(format string, v ...interface{})
}

type Authenticator interface {
  Auth(string, string) (uuid.UUID, error)
  GetUserByID(uuid.UUID) (*core.User, error)
  GetUserByEmail(string) (*core.User, error)
  GetUserByName(string) (*core.User, error)
  GetIdentities(uuid.UUID) ([]core.Identity, error)
  SetPassword(core.User, string) (bool, error)
}

type Mailer interface {
  SendPasswordTokenEmail(name string, email string, token uuid.UUID) error
  SendPasswordResetEmail(name string, email string) error
  SendRegistrationSuccessful(name string, email string) error
  SendUserRegistered(name string, email string) error
}

type Database interface {
  CreatePasswordResetToken(uuid.UUID) (uuid.UUID, error)
  ValidatePasswordToken(uuid.UUID, uuid.UUID) (bool, error)
  ScrubPasswordToken(uuid.UUID) error

  IsEmailUnique(string) (bool, error)
  IsNameUnique(string) (bool, error)

  AddPendingUser(name string, email string, template string, password string, summary string) error
}

type clientAuthResponse struct {
  Uuid uuid.UUID
  AccessLevel uint8
  Message string
  Success bool
}

type HttpConnector struct {
  store *sessions.CookieStore
  authenticator Authenticator
  logger Logger
  db Database
  mailer Mailer
}

func NewHttpConnector(sessionKey string, authenticator Authenticator, db Database, mailer Mailer, logger Logger) (*HttpConnector){
  gob.Register(uuid.UUID{})

  store := sessions.NewCookieStore([]byte(sessionKey))
  store.Options = &sessions.Options{
    Path: "/",
    MaxAge: 3600 * 8,
  }
  return &HttpConnector{store, authenticator, logger, db, mailer}
}

func (hc HttpConnector) LogoutHandler(w http.ResponseWriter, r *http.Request) {
  session, _ := hc.store.Get(r, "MGM")
  delete(session.Values, "guid")
  delete(session.Values,"address")
  session.Save(r,w)
  w.Header().Set("Content-Type", "application/json")
  w.Write([]byte("{\"Success\": true}"))
}

func (hc HttpConnector) ResumeHandler(w http.ResponseWriter, r *http.Request) {
  session, _ := hc.store.Get(r, "MGM")

  if len(session.Values) == 0 {
    response := clientAuthResponse{}
    js, err := json.Marshal(response)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/jons")
    w.Write(js)
    return
  }

  response := clientAuthResponse{session.Values["guid"].(uuid.UUID), session.Values["ulevel"].(uint8), "", true}
  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.Write(js)

}

func (hc HttpConnector) RegisterHandler(w http.ResponseWriter, r *http.Request) {

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

  if ! reg.Validate() {
    http.Error(w, "Invalid Request", http.StatusInternalServerError)
    return
  }

  /* Test if names are unique in pending users*/
  unique, err := hc.db.IsEmailUnique(reg.Email)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if ! unique {
    http.Error(w, "Credentials already exist", http.StatusInternalServerError)
    return
  }

  unique, err = hc.db.IsNameUnique(reg.Name)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if ! unique {
    http.Error(w, "Credentials already exist", http.StatusInternalServerError)
    return
  }

  /* Test if names are unique in registered users */
  user, err := hc.authenticator.GetUserByEmail(reg.Email)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if user != nil {
    http.Error(w, "Credentials already exist", http.StatusInternalServerError)
    return
  }
  user, err = hc.authenticator.GetUserByName(reg.Name)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if user != nil {
    http.Error(w, "Credentials already exist", http.StatusInternalServerError)
    return
  }

  err = hc.db.AddPendingUser(reg.Name, reg.Email, reg.Template, reg.Password, reg.Summary)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  _ = hc.mailer.SendRegistrationSuccessful(reg.Name, reg.Email)
  _ = hc.mailer.SendUserRegistered(reg.Name, reg.Email)

  w.Header().Set("Content-Type", "application/json")
  w.Write([]byte("{\"Success\": true}"))
}

func (hc HttpConnector) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
  type passwordReset struct {
    Name string
    Token uuid.UUID
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

  user, err := hc.authenticator.GetUserByName(req.Name)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if user == nil {
    http.Error(w, "Invalid Request, no user", http.StatusInternalServerError)
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

  isValid, err := hc.db.ValidatePasswordToken(user.UserID, req.Token)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if !isValid {
    http.Error(w, "Invalid Request, bad token", http.StatusInternalServerError)
    return
  }

  setPass, err := hc.authenticator.SetPassword(*user, req.Password)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if !setPass {
    http.Error(w, "Invalid Request, cant set", http.StatusInternalServerError)
    return
  }

  err = hc.db.ScrubPasswordToken(req.Token)
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

func (hc HttpConnector) PasswordTokenHandler(w http.ResponseWriter, r *http.Request) {
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

  user, err := hc.authenticator.GetUserByEmail(addr.Address)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if user == nil {
    http.Error(w, "Invalid Request, presence", http.StatusInternalServerError)
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

  token, err := hc.db.CreatePasswordResetToken(user.UserID)
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

func (hc HttpConnector) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

  guid,err := hc.authenticator.Auth(t.Username,t.Password);
  if err != nil {
    response := clientAuthResponse{Message: err.Error(),}
    js, err := json.Marshal(response)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
    return
  }

  user, err := hc.authenticator.GetUserByID(guid)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if user == nil {
    http.Error(w, "User account dissappeared", http.StatusInternalServerError)
    return
  }

  session, _ := hc.store.Get(r, "MGM")
  session.Values["guid"] = guid
  session.Values["address"] = r.RemoteAddr
  session.Values["ulevel"] = user.AccessLevel
  err = session.Save(r,w)
  if err != nil {
    hc.logger.Error("Error in httpConnector: %v", err)
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

  w.Header().Set("Content-Type", "application/jons")
  w.Write(js)
}
