package region

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// RestConsole is connection to a regions HTTP Rest console interface
type RestConsole interface {
	Write(string)
	Read() <-chan string
	Close()
}

// NewRestConsole constructs and connects a rest console
func NewRestConsole(r mgm.Region, h mgm.Host) (RestConsole, error) {
	c := console{
		read:    make(chan string, 1024),
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

	return c, nil
}

type console struct {
	url       string
	sessionID uuid.UUID
	read      chan string
	write     chan string
	closing   chan bool
}

func (c *console) connect(uname string, pass string) error {
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

func (c console) readProcess() {
	for {
		select {
		case <-c.closing:
			return
		default:
			//not closing, lets read
			resp, err := http.PostForm(c.url+"ReadResponses/"+c.sessionID.String()+"/", url.Values{"ID": {c.sessionID.String()}})
			if err != nil {
				c.read <- "Error opening console"
				continue
			}
			defer resp.Body.Close()

			if resp.ContentLength == 0 {
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				c.read <- "Error reading from console"
				return
			}

			type consoleConnectXML struct {
				XMLName xml.Name `xml:"ConsoleSession"`
				Lines   []string `xml:"Line"`
			}

			ss := consoleConnectXML{}

			err = xml.Unmarshal(body, &ss)
			if err != nil {
				c.read <- err.Error()
			}

			for _, line := range ss.Lines {
				c.read <- line
			}
		}
	}
}

func (c console) writeProcess() {
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
			if err != nil {
				c.read <- "Error writing to console"
			}
		}
	}
}

func (c console) Close() {
	http.PostForm(c.url+"CloseSession/", url.Values{"ID": {c.sessionID.String()}})
	close(c.closing)
}

func (c console) Read() <-chan string {
	return c.read
}

func (c console) Write(cmd string) {
	c.write <- cmd
}
