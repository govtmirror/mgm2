package simian

import (
  "net/url"
  "fmt"
  "github.com/satori/go.uuid"
)

func (sc simianConnector)EmailIsRegistered(email string) (exists bool, err error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "Email": {email},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)CreateUserEntry(username string, email string) (uuid.UUID, error){
  userID := uuid.NewV4()
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddUser"},
      "UserID": {userID.String()},
      "Name": {username},
      "Email": {email},
      "AccessLevel": {"0"},
    })
  
  if err != nil {
    return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return userID, nil
  }
  return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)CreateUserInventory(userID uuid.UUID, template string) (bool, error){
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddInventory"},
      "OwnerID": {userID.String()},
      "AvatarType": {template},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}