package simian

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Connector is an access point for communicating with a simiangrid installation
type Connector struct {
	url string
}

// NewConnector constructs a connector to communicate with Simian Grid
func NewConnector(simianURL string) (Connector, error) {
	sim := Connector{url: simianURL}

	//Test a connection from simianInstance to
	url := fmt.Sprintf("http://%v/Grid/", sim.url)
	resp, err := http.Get(url)
	if err != nil {
		return Connector{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Connector{}, err
	}

	result := string(body)

	if result != "SimianGrid" {
		return Connector{}, fmt.Errorf("Received %s instead of SimianGrid from Simian /Grid/ path", result)
	}

	sim.url = url
	return sim, nil
}

func (sc *Connector) handleRequest(remoteURL string, vals url.Values) ([]byte, error) {
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

func (sc Connector) confirmRequest(response []byte) error {
	var m confirmRequest
	err := json.Unmarshal(response, &m)
	if err != nil {
		return err
	}
	if m.Success {
		return nil
	}
	return fmt.Errorf("Error communicating with simian: %v", m.Message)
}
