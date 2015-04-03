package mgm

import (
    "fmt"
    "net"
    "os"
    "encoding/json"
)

type OpenSimListener struct {
    Port string
}

func (l* OpenSimListener) Listen() {
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

func (l* OpenSimListener) handleRequest(conn net.Conn){
    fmt.Println("New Connection Received")

    defer conn.Close()
    
    r := Region{} //Region is zeroed out
    
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
}