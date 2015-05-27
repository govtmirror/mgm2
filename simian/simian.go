package simian

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Connector is an interface to a Simian Grid installation
type Connector interface {
	Auth(username string, password string) (bool, uuid.UUID, error)
	GetUsers() ([]mgm.User, error)
	GetUserByID(id uuid.UUID) (mgm.User, error)
	GetUserByEmail(email string) (mgm.User, error)
	GetUserByName(name string) (mgm.User, error)
	GetIdentities(userID uuid.UUID) ([]core.Identity, error)
	SetPassword(userID uuid.UUID, password string) error
	ValidatePassword(userID uuid.UUID, password string) (bool, error)
	GetGroups() ([]mgm.Group, error)
	IsNameTaken(string) (bool, error)
	IsEmailTaken(string) (bool, error)
}

type simian struct {
	url string
}

// NewConnector constructs a connector to communicate with Simian Grid
func NewConnector(simianURL string) (Connector, error) {
	sim := simian{url: simianURL}

	//Test a connection from simianInstance to
	url := fmt.Sprintf("http://%v/Grid/", sim.url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := string(body)

	if result != "SimianGrid" {
		return nil, &errorString{fmt.Sprintf("Received %s instead of SimianGrid from Simian /Grid/ path", result)}
	}

	sim.url = url
	return sim, nil
}

type errorString struct {
	s string
}

func (es *errorString) Error() string {
	return es.s
}

func (sc *simian) handleRequest(remoteURL string, vals url.Values) ([]byte, error) {
	resp, err := http.PostForm(remoteURL, vals)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (sc simian) confirmRequest(response []byte) error {
	var m confirmRequest
	err := json.Unmarshal(response, &m)
	if err != nil {
		return err
	}
	if m.Success {
		return nil
	}
	return &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}
