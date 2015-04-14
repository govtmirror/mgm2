package webClient

import (
  "net/http"
  "encoding/json"
  "github.com/satori/go.uuid"
  "github.com/gorilla/sessions"
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
}

type notifier interface {

}

type database interface {

}

type HttpConnector struct {
  store *sessions.CookieStore
  authenticator Authenticator
  logger Logger
}

func NewHttpConnector(sessionKey string, authenticator Authenticator, logger Logger) (*HttpConnector){
  store := sessions.NewCookieStore([]byte(sessionKey))
  store.Options = &sessions.Options{
    Path: "/",
    MaxAge: 3600 * 8,
  }
  return &HttpConnector{store, authenticator, logger}
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
    hc.logger.Error("Invalid auth request")
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
