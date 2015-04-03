package simian

import (
  "net/url"
  "fmt"
  "github.com/satori/go.uuid"
)

func (sc simianConnector)GetGroups() ( map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"Group"},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Entries"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetGroupByID(groupID uuid.UUID) ( interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"Group"},
      "OwnerID": {groupID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    entries :=  m["Entries"].(map[string]interface{})
    var key string
    for key, _ = range entries {
      break
    }
    return entries[key], nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetGroupMembers(groupID uuid.UUID) ( map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"GroupMember"},
      "OwnerID": {groupID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Entries"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)GetGroupRoles(groupID uuid.UUID) ( map[string]interface{}, error) {
  m, err := sc.handle_request(simianInstance.url,
    url.Values{
      "RequestMethod": {"GetGenerics"},
      "Type": {"GroupRole"},
      "OwnerID": {groupID.String()},
    })
  
  if err != nil {
    return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
  }
  
  if _, ok := m["Success"]; ok {
    return m["Entries"].(map[string]interface{}), nil
  }
  return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m["Message"].(string))}
}

func (sc simianConnector)RemoveUserFromGroup(userID uuid.UUID, groupID uuid.UUID) ( bool, error) {
  //clear user role in group
  m, err := sc.handle_request(simianInstance.url,
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

func (sc simianConnector)AddUserToGroup(userID uuid.UUID, groupID uuid.UUID) ( bool, error) {
  m, err := sc.handle_request(simianInstance.url,
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
  m, err := sc.handle_request(simianInstance.url,
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
  m, err := sc.handle_request(simianInstance.url,
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

func (sc simianConnector)GetActiveGroup(userID uuid.UUID) ( uuid.UUID, error) {
  m, err := sc.handle_request(simianInstance.url,
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