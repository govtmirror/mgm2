package simian

import (
  "net/http"
  "net/url"
  "fmt"
  "io/ioutil"
  "encoding/json"
)

type simianConnector struct {
  url string
}

type errorString struct {
    s string
}
func (es *errorString) Error() string {
  return es.s
}

var simianInstance *simianConnector = nil

func (sc *simianConnector)handle_request(remoteUrl string, vals url.Values) (map[string]interface{}, error) {
  resp, err := http.PostForm(remoteUrl, vals)
  if err != nil {
    return nil, err
  }
  
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  
  m := map[string]interface{}{}
  err = json.Unmarshal(body, &m)
  if err != nil {
    return nil, err
  }
  
  return m, nil
}

func Instance() (*simianConnector, error) {
  if simianInstance == nil {
    return nil, &errorString{"simian has not been initialized"}
  }
  return simianInstance, nil
}

func Initialize(simianUrl string) (error) {
  if simianInstance == nil {
    simianInstance = &simianConnector{url: simianUrl}
  } else {
    return &errorString{"simian has already been initialized"}
  }
  //Test a connection from simianInstance to
  url := fmt.Sprintf("http://%v/Grid/", simianInstance.url)
  fmt.Println("simian testing on ", url)
  resp, err := http.Get(url)
  if err != nil {
    return err
  }
  
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return err
  }
  
  result := string(body)
  
  if result == "SimianGrid" {
    simianInstance.url = fmt.Sprintf("http://%v/Grid/",simianUrl);
    fmt.Println("simian initialized")
    return nil
  }
  return &errorString{fmt.Sprintf("Received %s instead of SimianGrid from Simian /Grid/ path")}
}
