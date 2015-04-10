package opensim

import (
  "fmt"
  "net"
  "os"
//  "encoding/json"
  "github.com/satori/go.uuid"
)

type Region interface {
}

type RegionManager interface {
  GetRegion(uuid.UUID) Region
}

type OpensimListener struct {
  Port string
  regionMgr RegionManager
}

func NewOpensimListener(port string, mgr RegionManager) (*OpensimListener, error){
  return &OpensimListener{port, mgr}, nil
}

func (l* OpensimListener) Listen() {
  link, err := net.Listen("tcp", ":"+l.Port)
  if err != nil {
    fmt.Println("Error Listening:", err.Error())
    os.Exit(1)
  }

  defer link.Close()
  fmt.Println("Listening for opensim on " + ":" + l.Port)
  for {
    conn, err := link.Accept()
    if err != nil {
      fmt.Println("Error accepting: ", err.Error())
      os.Exit(1)
    }

    go l.handleRequest(conn)
  }
}

func (l* OpensimListener) handleRequest(conn net.Conn){
  fmt.Println("New Connection Received")

  defer conn.Close()

  //we need some information from the region before we can process it
  /*
  r := l.regionMgr.GetRegion(uuid.UUID{})
    
  for {
    m := map[string]interface{}{}
    err := json.NewDecoder(conn).Decode(&m)
    if err != nil {
      fmt.Printf("Region %v went away: %v\n", r.name, err)
      r.Cleanup()
      return
    }
    switch m["type"] {
    case "frame":
      val := int(m["ms"].(float64))
      r.frames <- int(val)
    case "register":
      r.Register(m)
    default:
      fmt.Println(m)
    }      
  }
  */
}