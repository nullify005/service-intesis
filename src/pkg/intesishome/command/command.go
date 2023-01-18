package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"time"
)

const (
	_keepaliveUid  int           = 10
	_keepalive     time.Duration = 30 * time.Second
	_setTimeout    time.Duration = 15 * time.Second
	_authTimeout   time.Duration = 6 * time.Second
	_readyTimeout  time.Duration = 6 * time.Second
	_cmdStatus     string        = "status"
	_cmdSetAck     string        = "set_ack"
	_cmdSet        string        = "set"
	_cmdConnectAck string        = "connect_rsp"
	_cmdConnect    string        = "connect_req"
	_cmdConnectOk  string        = "ok"
	_cmdRssi       string        = "rssi"
	_cmdKeepalive  string        = "get"
)

type Command struct {
	logger         *log.Logger
	conn           net.Conn
	device         int64
	token          int
	keepalive      time.Duration
	shutdown       chan bool
	listenFinished chan bool
	setEvt         chan Response
	authEvt        chan Response
	readyEvt       chan bool
	err            error
	ready          bool
}

type Response struct {
	Command string `json:"command"`
	Data    struct {
		DeviceID int64  `json:"devceId,omitempty"`
		SeqNo    int    `json:"seqNo,omitempty"`
		Rssi     int    `json:"rssi,omitempty"`
		Status   string `json:"status,omitempty"`
	} `json:"data"`
}

type Request struct {
	Command string      `json:"command"`
	Data    RequestData `json:"data"`
}

type RequestData struct {
	DeviceID int64 `json:"deviceId"`
	Uid      int   `json:"uid"`
	Value    int   `json:"value"`
	SeqNo    int   `json:"seqNo"`
	Token    int   `json:"token"`
}

type Option func(*Command)

func WithLogger(l *log.Logger) Option {
	return func(c *Command) {
		c.logger = l
	}
}

func WithKeepalive(t time.Duration) Option {
	return func(c *Command) {
		c.keepalive = t
	}
}

func WithConn(n net.Conn) Option {
	return func(c *Command) {
		c.conn = n
	}
}

func New(opts ...Option) *Command {
	a := &Command{
		logger:         log.New(os.Stdout, "" /* prefix */, log.Ldate|log.Ltime|log.Lshortfile),
		shutdown:       make(chan bool, 1),
		listenFinished: make(chan bool, 1),
		readyEvt:       make(chan bool, 1),
		setEvt:         make(chan Response, 1),
		authEvt:        make(chan Response, 1),
		token:          0,
		keepalive:      _keepalive,
		err:            nil,
		ready:          false,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (c *Command) Connect(server string) (err error) {
	c.conn, err = net.Dial("tcp", server)
	return
}

func (c *Command) Token(token int) {
	c.token = token
}

func (c *Command) Set(device int64, uid, value int) error {
	if err := waitForReady(c); err != nil {
		return fmt.Errorf("set command failed. cause: %v", err)
	}
	cmd := &Request{
		Command: "set",
		Data: RequestData{
			DeviceID: device,
			Uid:      uid,
			Value:    value,
			SeqNo:    0,
		},
	}
	if err := socketWrite(c, cmd); err != nil {
		c.logger.Printf("set command failed. set: %+v cause: %v", cmd, err)
		return err
	}
	c.logger.Printf("sent set command: %+v", cmd)
	select {
	case <-time.After(_setTimeout):
		return fmt.Errorf("set command failed. cause: timeout waiting for set response")
	case set := <-c.setEvt:
		c.logger.Printf("set response received: %+v", set)
		return nil
	}
}

func (c *Command) Listen() error {
	go func() {
		err := c.auth()
		if err != nil {
			c.err = err
			c.shutdown <- true
		}
	}()
	eventLoop(c)
	return c.err
}

func (c *Command) Err() error {
	return c.err
}

func (c *Command) Close() {
	c.shutdown <- true
	<-c.listenFinished
}

func (c *Command) auth() error {
	c.logger.Print("authenticating ...")
	auth := &Request{
		Command: _cmdConnect,
		Data: RequestData{
			Token: c.token,
		},
	}
	if err := socketWrite(c, auth); err != nil {
		c.logger.Printf("auth socket write error. cause: %v", err)
		return err
	}
	select {
	case <-time.After(_authTimeout):
		return fmt.Errorf("timeout waiting for authentication response")
	case auth := <-c.authEvt:
		if auth.Data.Status != _cmdConnectOk {
			c.logger.Printf("invalid auth response. expected: %s got: %s", _cmdConnectOk, auth.Data.Status)
			c.Close()
			return fmt.Errorf("invalid auth response")
		}
	}
	c.logger.Print("authentication successful")
	c.ready = true
	return nil
}

func eventLoop(c *Command) {
	defer c.conn.Close()
	keepaliveEvt := time.NewTicker(c.keepalive)
	socketEvt := make(chan string, 1)
	readerErr := make(chan error, 1)
	done := make(chan bool, 1)
	go socketReader(c, socketEvt, readerErr)
	go func() {
		c.logger.Print("event listener starting")
	eventLoop:
		for {
			select {
			case <-c.shutdown:
				c.logger.Print("shutdown received")
				break eventLoop
			case <-keepaliveEvt.C:
				c.logger.Print("sending keepalive")
				if err := keepalive(c); err != nil {
					c.logger.Printf("error sending keepalive. cause: %v", err)
					c.err = err
					break eventLoop
				}
			case err := <-readerErr:
				c.logger.Printf("received reader error: %v", err)
				c.err = err
				break eventLoop
			case buf := <-socketEvt:
				c.logger.Printf("socket event received: %s", buf)
				cmdResp := &Response{}
				if err := json.Unmarshal([]byte(buf), &cmdResp); err != nil {
					c.logger.Printf("response decode error. resp: %s cause: %v", buf, err)
				} else {
					go cmdHandler(c, cmdResp)
				}
			}
		}
		done <- true
	}()
	<-done
	c.listenFinished <- true
	c.ready = false
	c.logger.Print("event listener stopped")
}

func cmdHandler(c *Command, r *Response) {
	switch r.Command {
	case _cmdRssi:
		c.logger.Printf("received rssi update. %+v", r)
	case _cmdStatus:
		c.logger.Printf("received status. %+v", r)
	case _cmdSetAck:
		c.logger.Printf("received ack. %+v", r)
		// TODO: should we be calling status to confirm that it's actually true?
		c.setEvt <- *r
	case _cmdConnectAck:
		c.logger.Printf("received connect. %+v", r)
		c.authEvt <- *r
	default:
		c.logger.Printf("unhandled command. %+v", r)
	}
}

func socketReader(c *Command, outputC chan string, errC chan error) {
	scanner := bufio.NewScanner(c.conn)
	scanner.Split(jsonSplit)
	for {
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			c.logger.Printf("scanner error. cause: %v", err)
			errC <- err
			return
		}
		buf := scanner.Bytes()
		buf = bytes.Trim(buf, "\x00") // trim any nulls from the end
		c.logger.Printf("scanner received: %s", buf)
		outputC <- string(buf)
	}
}

func keepalive(c *Command) error {
	cmdReq := &Request{
		Command: _cmdKeepalive,
		Data: RequestData{
			DeviceID: c.device,
			Uid:      _keepaliveUid,
		},
	}
	if err := socketWrite(c, cmdReq); err != nil {
		c.logger.Printf("keepalive socket write error. cause: %v", err)
		return err
	}
	return nil
}

func socketWrite(c *Command, thing interface{}) error {
	b, err := json.Marshal(thing)
	if err != nil {
		e := fmt.Sprintf("json encode error. auth: %+v cause: %v", thing, err)
		c.logger.Print(e)
		return fmt.Errorf(e)
	}
	length, err := c.conn.Write(b)
	if err != nil {
		e := fmt.Sprintf("socket write error. cause: %v", err)
		c.logger.Print(e)
		return fmt.Errorf(e)
	}
	if length != len(b) {
		e := fmt.Sprintf("write length mismatch. %d != %d", length, len(b))
		c.logger.Print(e)
		return fmt.Errorf(e)
	}
	return nil
}

func jsonSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	l := log.New(os.Stdout, "" /* prefix */, log.Ldate|log.Ltime|log.Lshortfile)
	if atEOF && len(data) == 0 {
		l.Print("at EOF with no data")
		return 0, nil, nil
	}

	start := -1
	opens := 0
	closes := 0
	l.Printf("processing: %s", data)
	for i := 0; i < len(data); i++ {
		if data[i] == '{' {
			if start == -1 {
				start = i
			}
			opens++
		}
		if data[i] == '}' {
			closes++
			if opens == closes {
				l.Printf("found json payload: %s", data[start:i+1])
				return i + 1, data[start : i+1], nil
			}
		}
	}

	if atEOF {
		l.Printf("at EOF with data: %s", data)
		return len(data), data, nil
	}
	return 0, nil, nil
}

func doubleBraceSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	r := regexp.MustCompile("}}")
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	loc := r.FindIndex(data)
	if loc != nil && loc[0] >= 0 {
		return loc[1], data[:loc[1]], nil
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func waitForReady(c *Command) error {
	for tries := 1; tries <= 3; tries++ {
		if c.ready {
			return nil
		}
		c.logger.Printf("waiting for connection to be ready (%d tries) ...", tries)
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("timeout waiting for connection to be ready")
}
