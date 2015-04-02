package simian

import (
  "net/http"
  "fmt"
  "io/ioutil"
)

type SimianResponse struct {
}

type SimianRequest struct {
  callback chan SimianResponse
}

type simianConnector struct {
  requests chan SimianRequest
  url string
}

type errorString struct {
    s string
}
func (es *errorString) Error() string {
  return es.s
}

var simianInstance *simianConnector = nil

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
    fmt.Println("simian initialized")
    return nil
  }
  return &errorString{fmt.Sprintf("Received %s instead of SimianGrid from Simian /Grid/ path")}
}
