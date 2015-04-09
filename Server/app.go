package main

import (
  "github.com/M-O-S-E-S/mgm2/mgm"
  "fmt"
  "os"
  "encoding/json"
)

//global config object
type config struct {
}

func main() {
  mgmConfig := mgm.MgmConfig{}
  conf := config{}
  
  fmt.Println("Reading configuration file")
  file, _ := os.Open("conf.json")
  decoder := json.NewDecoder(file)

  err := decoder.Decode(
    &struct{ 
      *mgm.MgmConfig
      *config
    }{&mgmConfig,&conf})
  if err != nil {
    fmt.Println("Error readig config file: ", err)
    return
  }
  
  mgmCore, err := mgm.NewMGM(mgmConfig)
  if err != nil {
    fmt.Println("Error instantiating MGMCore ", err)
    return
  }

  mgmCore.Listen()
}