package webClient

import (
  "strings"
  "net/mail"
)

type registrant struct {
  Name string
  Email string
  Password string
  Template string
  Summary string
}

func (r registrant) Validate() bool {
  if r.Name == "" {
    return false
  }
  names := strings.Split(r.Name, " ")
  if len(names) != 2 {
    return false
  }

  if r.Email == "" {
    return false
  }
  _, err := mail.ParseAddress(r.Email)
  if err != nil {
    return false
  }

  if r.Password == "" {
    return false
  }

  if r.Template != "M" && r.Template != "F" {
    return false
  }
  return true
}
