package core

import (
  "encoding/json"
)


type userRequest struct {
  MessageType string
  Message json.RawMessage
}

func (ur *userRequest)load(msg []byte) {
  err := json.Unmarshal(msg, ur)
  if err != nil {
    ur.MessageType = err.Error()
  }
}