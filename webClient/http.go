package webClient

import (
  "net/http"
  "net/mail"
  "encoding/json"
  "github.com/satori/go.uuid"
  "github.com/gorilla/sessions"
   "github.com/M-O-S-E-S/mgm2/core"
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
  GetUserByEmail(string) (core.User, error)
  GetIdentities(uuid.UUID) ([]core.Identity, error)
}

type Mailer interface {
  SendPasswordResetEmail(name string, email string, token uuid.UUID) error
}

type Database interface {
  CreatePasswordResetToken(uuid.UUID) (uuid.UUID, error)
}

type HttpConnector struct {
  store *sessions.CookieStore
  authenticator Authenticator
  logger Logger
  db Database
  mailer Mailer
}

func NewHttpConnector(sessionKey string, authenticator Authenticator, db Database, mailer Mailer, logger Logger) (*HttpConnector){
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
    
  type clientAuthResponse struct {
    Uuid string
    Success bool
  }
  
  if len(session.Values) == 0 {
    response := clientAuthResponse{"", false}
    js, err := json.Marshal(response)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/jons")
    w.Write(js)
    return
  }
  
  response := clientAuthResponse{session.Values["guid"].(string), true}
  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  w.Header().Set("Content-Type", "application/json")
  w.Write(js)

}

func (hc HttpConnector) RegisterHandler(w http.ResponseWriter, r *http.Request) {
  
}

func (hc HttpConnector) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
  
}

func (hc HttpConnector) PasswordTokenHandler(w http.ResponseWriter, r *http.Request) {
  type emailRequest struct {
    Email string
  }

  if r.Method != "POST" {
    http.Error(w, "Invalid Request", http.StatusInternalServerError)
    return
  }

  var req emailRequest
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

  hc.logger.Info("Email validated")

  user, err := hc.authenticator.GetUserByEmail(addr.Address)
  if err != nil {
    http.Error(w, "Invalid Request, presence", http.StatusInternalServerError)
    return
  }

  hc.logger.Info("user found")

  ids, err := hc.authenticator.GetIdentities(user.UserID)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  hc.logger.Info("Identities received")

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
  err = hc.mailer.SendPasswordResetEmail(user.Name, user.Email, token)
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
  
  type clientAuthResponse struct {
    Uuid uuid.UUID
    Message string
    Success bool
  }
  
  guid,err := hc.authenticator.Auth(t.Username,t.Password);
  if err != nil {
    response := clientAuthResponse{uuid.UUID{}, err.Error(), false}
    js, err := json.Marshal(response)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
    return
  }
  
  response := clientAuthResponse{guid, "", true}
  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  session, _ := hc.store.Get(r, "MGM")
  session.Values["guid"] = guid.String()
  session.Values["address"] = r.RemoteAddr
  err = session.Save(r,w)
  if err != nil {
    hc.logger.Error("Error in httpConnector: %v", err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  hc.logger.Info("Session saved, returning success")
  
  w.Header().Set("Content-Type", "application/jons")
  w.Write(js)
}
