package simian

type SimianResponse struct {
}

type SimianRequest struct {
  callback chan SimianResponse
}

type simianConnector struct {
  requests chan SimianRequest
}

var simianInstance *simianConnector = nil

func Instance() *simianConnector {
  if simianInstance == nil {
    simianInstance = new(simianConnector)
  }
  return simianInstance
}

