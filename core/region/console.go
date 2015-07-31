package region

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// NewRestConsole constructs and connects a rest console
func NewRestConsole(r mgm.Region, h mgm.Host) (RestConsole, error) {
	c := RestConsole{
		read:    make(chan []string, 36),
		write:   make(chan string, 8),
		closing: make(chan bool),
	}

	c.url = fmt.Sprintf("http://%v:%v/", h.Address, r.ConsolePort)

	err := c.connect(r.ConsoleUname.String(), r.ConsolePass.String())
	if err != nil {
		return c, err
	}

	go c.readProcess()
	go c.writeProcess()

	c.initialized = true

	return c, nil
}

//RestConsole is an object representing a rest console connection with a remote process
type RestConsole struct {
	url         string
	sessionID   uuid.UUID
	read        chan []string
	write       chan string
	closing     chan bool
	initialized bool
}

//IsConnected is a simple test if the console is active or not
func (c RestConsole) IsConnected() bool {
	return c.initialized
}

func (c *RestConsole) connect(uname string, pass string) error {
	resp, err := http.PostForm(c.url+"StartSession/", url.Values{"USER": {uname}, "PASS": {pass}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type consoleConnectXML struct {
		XMLName   xml.Name  `xml:"ConsoleSession"`
		SessionID uuid.UUID `xml:"SessionID"`
		Prompt    string
	}

	ss := consoleConnectXML{}
	err = xml.Unmarshal(body, &ss)
	if err != nil {
		return err
	}
	c.sessionID = ss.SessionID
	return nil
}

func (c RestConsole) readProcess() {
	for {
		select {
		case <-c.closing:
			return
		default:
			//not closing, lets read
			resp, err := http.PostForm(c.url+"ReadResponses/"+c.sessionID.String()+"/", url.Values{"ID": {c.sessionID.String()}})
			if err != nil {
				c.read <- []string{"Error opening console"}
				continue
			}
			defer resp.Body.Close()

			if resp.ContentLength == 0 {
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.read <- []string{"Error reading from console"}
				return
			}

			type consoleConnectXML struct {
				XMLName xml.Name `xml:"ConsoleSession"`
				Lines   []string `xml:"Line"`
			}

			ss := consoleConnectXML{}

			err = xml.Unmarshal(body, &ss)
			if err != nil {
				c.read <- []string{err.Error()}
			}

			c.read <- ss.Lines
		}
	}
}

func (c RestConsole) writeProcess() {
	for {
		select {
		case <-c.closing:
			return
		case cmd := <-c.write:
			_, err := http.PostForm(
				c.url+"SessionCommand/",
				url.Values{
					"ID":      {c.sessionID.String()},
					"COMMAND": {cmd},
				},
			)
			timestamp := time.Now()
			h, m, s := timestamp.Clock()
			c.read <- []string{fmt.Sprintf("0:normal:%v:%v:%v - %v", h, m, s, cmd)}
			if err != nil {
				c.read <- []string{"Error writing to console"}
			}
		}
	}
}

//Close closes a rest console session with a remote instance
func (c *RestConsole) Close() {
	if c.initialized {
		http.PostForm(c.url+"CloseSession/", url.Values{"ID": {c.sessionID.String()}})
		close(c.closing)
		c.initialized = false
	}
}

func (c RestConsole) Read() <-chan []string {
	return c.read
}

func (c RestConsole) Write(cmd string) {
	if c.initialized {
		c.write <- cmd
	}
}
