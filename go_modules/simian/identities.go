package simian

import (
  "net/url"
  "crypto/md5"
  "fmt"
  "encoding/hex"
  "github.com/satori/go.uuid"
)

/* Example using Auth
  inst, _ := simian.Instance()
  uuid, err := inst.Auth("test load_9", "password123")
  if err != nil {
    fmt.Println("Error authorizing: ", err);
    return
  }
  fmt.Println(uuid)
*/

func (sc simianConnector)Auth(username string, password string) (uid uuid.UUID, err error) {
  hasher := md5.New()
  hasher.Write([]byte(password))
  
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AuthorizeIdentity"}, 
      "Identifier": {username},
      "Credential": {hex.EncodeToString(hasher.Sum(nil))},
      "Type": {"md5hash"},
    })
  
  if err != nil {
    return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    userID, _ := uuid.FromString(m["UserID"].(string))
    return  userID, nil
  }
  return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)EnableIdentity(username string, identityType string, credential string, userID uuid.UUID) (bool, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddIdentity"},
      "Identifier": {username},
      "Type": {identityType},
      "Credential": {credential},
      "UserID": {userID.String()},
      "Enabled": {"true"},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)DisableIdentity(username string, identityType string, credential string, userID uuid.UUID) (bool, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddIdentity"},
      "Identifier": {username},
      "Type": {identityType},
      "Credential": {credential},
      "UserID": {userID.String()},
      "Enabled": {"false"},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)InsertPasswordHash(username string, credential string, userID uuid.UUID) (bool, error) {
  m, err := sc.handle_request(simianInstance.url,
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
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)SetPassword(username string, password string, userID uuid.UUID) (bool, error) {
  hasher := md5.New()
  hasher.Write([]byte(password))
  
  m, err := sc.handle_request(simianInstance.url,
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
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetIdentities(userID uuid.UUID) ( map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetIdentities"},
      "UserID": {userID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Identities"].( map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

