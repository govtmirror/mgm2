package simian

import (
  "net/url"
  "crypto/md5"
  "fmt"
  "encoding/hex"
  "encoding/json"
  "github.com/satori/go.uuid"
  "github.com/M-O-S-E-S/mgm2/core"
)

func (sc SimianConnector)Auth(username string, password string) (uuid.UUID, error) {
  hasher := md5.New()
  hasher.Write([]byte(password))
  
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AuthorizeIdentity"}, 
      "Identifier": {username},
      "Credential": {hex.EncodeToString(hasher.Sum(nil))},
      "Type": {"md5hash"},
    })
  
  if err != nil {
    return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  type tmpStruct struct {
    Success bool
    Message string
    UserID string
  }
  var m tmpStruct
  err = json.Unmarshal(response, &m)
  if err != nil {
    return uuid.UUID{}, err
  }
  if m.Success {
    userID, _ := uuid.FromString(m.UserID)
    return  userID, nil
  }
  return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc SimianConnector)EnableIdentity(username string, identityType string, credential string, userID uuid.UUID) (bool, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddIdentity"},
      "Identifier": {username},
      "Type": {identityType},
      "Credential": {credential},
      "UserID": {userID.String()},
      "Enabled": {"1"},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)DisableIdentity(username string, identityType string, credential string, userID uuid.UUID) (bool, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddIdentity"},
      "Identifier": {username},
      "Type": {identityType},
      "Credential": {credential},
      "UserID": {userID.String()},
      "Enabled": {"0"},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)InsertPasswordHash(username string, credential string, userID uuid.UUID) (bool, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddIdentity"},
      "Identifier": {username},
      "Type": {"md5hash"},
      "Credential": {credential},
      "UserID": {userID.String()},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)SetPassword(username string, password string, userID uuid.UUID) (bool, error) {
  hasher := md5.New()
  hasher.Write([]byte(password))
  
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddIdentity"},
      "Identifier": {username},
      "Type": {"md5hash"},
      "Credential": {hex.EncodeToString(hasher.Sum(nil))},
      "UserID": {userID.String()},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)GetIdentities(userID uuid.UUID) ( []core.Identity, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetIdentities"},
      "UserID": {userID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  type req struct {
    Success bool
    Message string
    Identities []core.Identity
  }
  var m req
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    return  m.Identities, nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}
