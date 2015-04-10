package simian

import (
  "net/url"
  "fmt"
  "github.com/satori/go.uuid"
  "encoding/json"
)

func (sc SimianConnector)GetGroups() ( []Group, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"Group"},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  type req struct {
    Success bool
    Message string
    Entries []Group
  }
  
  var m req
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    return  m.Entries, nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc SimianConnector)GetGroupByID(groupID uuid.UUID) ( Group, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"Group"},
      "OwnerID": {groupID.String()},
    })
  
  if err != nil {
    return Group{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  type req struct {
    Success bool
    Message string
    Entries []Group
  }
  
  var m req
  err = json.Unmarshal(response, &m)
  if err != nil {
    return Group{}, err
  }
  if m.Success {
    return  m.Entries[0], nil
  }
  return Group{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc SimianConnector)GetGroupMembers(groupID uuid.UUID) ( []uuid.UUID, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"GroupMember"},
      "Key": {groupID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  type req struct {
    Success bool
    Message string
    Entries []Generic
  }
  
  var m req
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    users := make([]uuid.UUID, len(m.Entries))
    for index, el := range m.Entries {
      users[index] = el.OwnerID
    }
    return  users, nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc SimianConnector)GetGroupRoles(groupID uuid.UUID) ( []string, error) {
  response, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"GroupRole"},
      "OwnerID": {groupID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  type req struct {
    Success bool
    Message string
    Entries []Generic
  }
  
  var m req
  err = json.Unmarshal(response, &m)
  if err != nil {
    return nil, err
  }
  if m.Success {
    roles := make([]string, len(m.Entries))
    for index, el := range m.Entries {
      roles[index] = el.Value
    }
    return  roles, nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}
/*
func (sc SimianConnector)RemoveUserFromGroup(userID uuid.UUID, groupID uuid.UUID) ( bool, error) {
  //clear user role in group
  m, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"RemoveGeneric"},
      "Type": {fmt.Sprintf("GroupRole%v", groupID.String())},
      "OwnerID": {userID.String()},
      "Key": {uuid.UUID{}.String()},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ! ok {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
  }
  
  // remove default role
  m, err = sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"RemoveGeneric"},
      "Type": {"GroupMember"},
      "OwnerID": {userID.String()},
      "Key": {groupID.String()},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ! ok {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
  }
  //purge active group record if it is this group
  activeGroup, err := sc.GetActiveGroup(userID);
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  if activeGroup == groupID {
    m, err = sc.handle_request(simianInstance.url,
      url.Values{
        "RequestMethod": {"RemoveGeneric"},
        "Type": {"Group"},
        "OwnerID": {userID.String()},
        "Key": {"ActiveGroup"},
      })

    if err != nil {
      return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
    }

    if _, ok := m["Success"]; ! ok {
      return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
    }
    
    //blank out active group now that the user has none
     m, err = sc.handle_request(simianInstance.url,
      url.Values{
        "RequestMethod": {"AddGeneric"},
        "Type": {"Group"},
        "OwnerID": {userID.String()},
        "Key": {"ActiveGroup"},
        "Value": {"{}"},
      })

    if err != nil {
      return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
    }

    if _, ok := m["Success"]; ! ok {
      return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
    }
  }
  return true, nil
}

func (sc SimianConnector)AddUserToGroup(userID uuid.UUID, groupID uuid.UUID) ( bool, error) {
  m, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"AddGeneric"},
      "OwnerID": {userID.String()},
      "Type": {"GroupMember"},
      "Key": {groupID.String()},
      "Value": {"{\"AcceptNotices\":true,\"ListInProfile\":false}"},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; !ok {
    return false, nil
  }
  
  m, err = sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"AddGeneric"},
      "OwnerID": {userID.String()},
      "Type": {"GroupRole"},
      "Key": {uuid.UUID{}.String()},
      "Value": {"{}"},
    })
  
  if err != nil {
    return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; !ok {
    return false, nil
  }
  
  return true, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetRolesForUser(userID uuid.UUID, groupID uuid.UUID) ( map[string]interface{}, error) {
  m, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {fmt.Sprintf("GroupRole%v",groupID.String())},
      "OwnerID": {userID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Entries"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetGroupsForUser(userID uuid.UUID) ( map[string]interface{}, error) {
  m, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"GroupMember"},
      "OwnerID": {userID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Entries"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc SimianConnector)GetActiveGroup(userID uuid.UUID) ( uuid.UUID, error) {
  m, err := sc.handle_request(sc.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"Group"},
      "OwnerID": {userID.String()},
    })
  
  if err != nil {
    return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    entries := m["Entries"].(map[string]interface{})
    if len(m) == 0 {
      return uuid.UUID{}, &errorString{fmt.Sprintf("User has no active groups")}
    }
    var key string
    for key, _ = range entries {
      break
    }
    entry := entries[key].(map[string]interface{})
    if groupid, ok := entry["GroupID"] ; ok {
      ag, _ := uuid.FromString(groupid.(string))
      return ag, nil
    }
  }
  return uuid.UUID{}, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}
*/