package simian

import (
  "net/url"
  "fmt"
  "github.com/satori/go.uuid"
)

func (sc simianConnector)GetUserByEmail(email string) (map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "Email": {email},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["User"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetUserByName(name string) (map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "Name": {name},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["User"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetUserByID(id uuid.UUID) (map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "UserID": {id.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["User"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetUsers() ( map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetUsers"},
      "NameQuery": {""},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Users"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)RemoveUser(userID uuid.UUID) ( bool, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"RemoveUser"},
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

func (sc simianConnector)SetUserLastLocation(userID uuid.UUID, uri string) ( bool, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddUserData"},
      "UserID": {userID.String()},
      "LastLocation": {uri},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)SetUserHome(userID uuid.UUID, uri string) ( bool, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddUserData"},
      "UserID": {userID.String()},
      "HomeLocation": {uri},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)UpdateUser(name string, email string, userID uuid.UUID, level int) ( bool, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddUser"},
      "UserID": {userID.String()},
      "Email": {email},
      "Name": {name},
      "AccessLevel": {fmt.Sprintf("%v", level)},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}