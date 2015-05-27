package simian

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

func (sc simian) GetGroups() ([]mgm.Group, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetGenerics"},
			"Type":          {"Group"},
		})

	if err != nil {
		return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	type simianGroupValues struct {
		GroupID       uuid.UUID
		ShowInList    bool
		InsigniaID    uuid.UUID
		FounderID     uuid.UUID
		EveronePowers []int
		OwnerRoleID   uuid.UUID
		OwnersPowers  []int
	}

	type simianGroup struct {
		OwnerID uuid.UUID
		Name    string `json:"Key"`
		Value   string
	}

	type req struct {
		Success bool
		Message string
		Entries []simianGroup
	}

	var m req
	err = json.Unmarshal(response, &m)
	if err != nil {
		return nil, err
	}
	if m.Success {
		blankID := uuid.UUID{}
		groups := make([]mgm.Group, 0)
		for _, sg := range m.Entries {
			sgv := simianGroupValues{}
			err = json.Unmarshal([]byte(sg.Value), &sgv)
			if err != nil {
				fmt.Println(err)
			}
			if sgv.FounderID == blankID {
				continue
			}
			group := mgm.Group{Members: make([]uuid.UUID, 0), Roles: make([]string, 0)}
			//map simian data into core.Group struct
			group.Name = sg.Name
			group.Founder = sgv.FounderID
			group.ID = sg.OwnerID
			//pre-populate Members and Roles
			members, _ := sc.GetGroupMembers(group.ID)
			group.Members = members
			roles, _ := sc.GetGroupRoles(group.ID)
			group.Roles = roles
			groups = append(groups, group)
		}
		return groups, nil
	}
	return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc simian) GetGroupByID(groupID uuid.UUID) (*mgm.Group, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetGenerics"},
			"Type":          {"Group"},
			"OwnerID":       {groupID.String()},
		})

	if err != nil {
		return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", err)}
	}

	type req struct {
		Success bool
		Message string
		Entries []mgm.Group
	}

	var m req
	err = json.Unmarshal(response, &m)
	if err != nil {
		return nil, err
	}
	if m.Success {
		return &m.Entries[0], nil
	}
	return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc simian) GetGroupMembers(groupID uuid.UUID) ([]uuid.UUID, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetGenerics"},
			"Type":          {"GroupMember"},
			"Key":           {groupID.String()},
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
		return users, nil
	}
	return nil, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}

func (sc simian) GetGroupRoles(groupID uuid.UUID) ([]string, error) {
	response, err := sc.handleRequest(sc.url,
		url.Values{
			"RequestMethod": {"GetGenerics"},
			"Type":          {"GroupRole"},
			"OwnerID":       {groupID.String()},
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
		return roles, nil
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
