package simian

import (
  "net/http"
  "net/url"
  "fmt"
  "io/ioutil"
  "encoding/json"
)

type SimianConnector struct {
  url string
}

func NewSimianConnector(simianUrl string) (*SimianConnector, error) {
  sim := &SimianConnector{url: simianUrl}
  
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
    return nil, &errorString{fmt.Sprintf("Received %s instead of SimianGrid from Simian /Grid/ path")}
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

func (sc *SimianConnector)handle_request(remoteUrl string, vals url.Values) ([]byte, error) {
  resp, err := http.PostForm(remoteUrl, vals)
  if err != nil {
    return nil, err
  }
  
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  
  return body, nil
}

func (sc SimianConnector)confirmRequest(response []byte) (bool, error){
  var m confirmRequest
  err := json.Unmarshal(response, &m)
  if err != nil {
    return false, err
  }
  if m.Success {
    return  true, nil
  }
  return false, &errorString{fmt.Sprintf("Error communicating with simian: %v", m.Message)}
}
