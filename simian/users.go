package simian

import (
  "net/url"
  "fmt"
  "github.com/satori/go.uuid"
  "encoding/json"
  "github.com/M-O-S-E-S/mgm2/core"
)

func (sc SimianConnector)GetUserByEmail(email string) (*core.User, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "Email": {email},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }

  var m userRequest
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    return  &m.User, nil
  }
  return nil, nil
}

func (sc SimianConnector)GetUserByName(name string) (*core.User, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "Name": {name},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  var m userRequest
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    return  &m.User, nil
  }
  return nil, nil
}

func (sc SimianConnector)GetUserByID(id uuid.UUID) (*core.User, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetUser"},
      "UserID": {id.String()},
    })

  var m userRequest
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    return  &m.User, nil
  }
  return nil, nil
}

func (sc SimianConnector)GetUsers() ( []core.User, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetUsers"},
      "NameQuery": {""},
    })
  
  var m usersRequest
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if !m.Success {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
  }
  //lookup suspension status for each user
  users := m.Users
  for idx, user := range users {
    ids, err := sc.GetIdentities(user.UserID)
    if err != nil {
      continue
    }
    isActive := false
    for _, id := range ids {
      if id.Enabled {
        isActive = true
      }
    }
    if !isActive {
      users[idx].Suspended = true
    }
  }
  return users, nil
}

func (sc SimianConnector)RemoveUser(userID uuid.UUID) ( bool, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"RemoveUser"},
      "UserID": {userID.String()},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)SetUserLastLocation(userID uuid.UUID, uri string) ( bool, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddUserData"},
      "UserID": {userID.String()},
      "LastLocation": {uri},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)SetUserHomeLocation(userID uuid.UUID, uri string) ( bool, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddUserData"},
      "UserID": {userID.String()},
      "HomeLocation": {uri},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  return sc.confirmRequest(response)
}

func (sc SimianConnector)UpdateUser(name string, email string, userID uuid.UUID, level int) ( bool, error) {
  response, err := sc.handle_request(sc.url,
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
  
  return sc.confirmRequest(response)
}
