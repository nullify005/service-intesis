package async

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

	"github.com/nullify005/service-intesis/pkg/intesishome"
)

type AsyncSocket struct {
	logger    *log.Logger
	conn      net.Conn
	device    int64
	keepalive time.Duration
	shutdown  chan bool
	setEvt    chan intesishome.CommandResponse
	authEvt   chan intesishome.CommandResponse
}

const (
	_readLimitBytes int           = 1024
	_keepalive      time.Duration = 30 * time.Second
	_keepaliveUid   int           = 10
	_setTimeout     time.Duration = 3 * time.Second
	_authTimeout    time.Duration = 3 * time.Second
	_cmdStatus      string        = "status"
	_cmdSetAck      string        = "set_ack"
	_cmdConnect     string        = "connect_rsp"
	_cmdConnectOk   string        = "ok"
	_cmdRssi        string        = "rssi"
)

type AsyncSocketOption func(*AsyncSocket)

func WithLogger(l *log.Logger) AsyncSocketOption {
	return func(a *AsyncSocket) {
		a.logger = l
	}
}

func WithDevice(d int64) AsyncSocketOption {
	return func(a *AsyncSocket) {
		a.device = d
	}
}

func WithKeepalive(t time.Duration) AsyncSocketOption {
	return func(a *AsyncSocket) {
		a.keepalive = t
	}
}

func New(opts ...AsyncSocketOption) *AsyncSocket {
	a := &AsyncSocket{
		logger:    log.New(os.Stdout, "" /* prefix */, log.Ldate|log.Ltime|log.Lshortfile),
		shutdown:  make(chan bool, 1),
		setEvt:    make(chan intesishome.CommandResponse, 1),
		authEvt:   make(chan intesishome.CommandResponse, 1),
		device:    127934703953,
		keepalive: _keepalive,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *AsyncSocket) Connect(place string) (err error) {
	a.conn, err = net.Dial("tcp", place)
	return
}

func (a *AsyncSocket) Auth(token int) (err error) {
	auth := &intesishome.CommandRequest{
		Command: "connect_req",
		Data: intesishome.CommandRequestData{
			Token: token,
		},
	}
	if err := socketWrite(a, auth); err != nil {
		a.logger.Printf("auth socket write error. cause: %v", err)
		return err
	}
	select {
	case <-time.After(_authTimeout):
		return fmt.Errorf("timeout waiting for authentication response")
	case auth := <-a.authEvt:
		if auth.Data.Status != _cmdConnectOk {
			a.conn.Close()
			return fmt.Errorf("invalid auth response. expected: %s got: %s", _cmdConnectOk, auth.Data.Status)
		}
	}
	a.logger.Print("authentication successful")
	return nil
}

func (a *AsyncSocket) Set(uid, value int) error {
	cmd := &intesishome.CommandRequest{
		Command: "set",
		Data: intesishome.CommandRequestData{
			DeviceID: a.device,
			Uid:      uid,
			Value:    value,
			SeqNo:    0,
		},
	}
	if err := socketWrite(a, cmd); err != nil {
		a.logger.Printf("set command failed. set: %+v cause: %v", cmd, err)
		return err
	}
	a.logger.Printf("sent set command: %+v", cmd)
	select {
	case <-time.After(_setTimeout):
		return fmt.Errorf("timeout waiting for set response")
	case set := <-a.setEvt:
		a.logger.Printf("set response received: %+v", set)
		return nil
	}
}

func (a *AsyncSocket) EventListener(token int) {
	go eventLoop(a)
	err := a.Auth(token)
	if err != nil {
		a.shutdown <- true
	}
}

func eventLoop(a *AsyncSocket) {
	defer a.conn.Close()
	keepaliveEvt := time.NewTicker(a.keepalive)
	socketEvt := make(chan string)
	go socketReader(a, socketEvt)
	go func() {
		a.logger.Print("starting event listener")
		for {
			select {
			case <-a.shutdown:
				a.logger.Print("shutdown received")
				return
			case <-keepaliveEvt.C:
				a.logger.Print("sending keepalive")
				if err := keepalive(a); err != nil {
					a.logger.Printf("error sending keepalive. cause: %v", err)
					a.shutdown <- true
				}
			case buffer := <-socketEvt:
				cmdResp := &intesishome.CommandResponse{}
				if err := json.Unmarshal([]byte(buffer), &cmdResp); err != nil {
					a.logger.Printf("response decode error. resp: %s cause: %v", buffer, err)
				} else {
					go cmdHandler(a, cmdResp)
				}
			}
		}
	}()
	<-a.shutdown
	a.logger.Print("shutdown")
}

func (a *AsyncSocket) Close() {
	a.shutdown <- true
}

func cmdHandler(a *AsyncSocket, c *intesishome.CommandResponse) {
	switch c.Command {
	case _cmdRssi:
		a.logger.Printf("received rssi update. %+v", c)
	case _cmdStatus:
		a.logger.Printf("received status. %+v", c)
	case _cmdSetAck:
		a.logger.Printf("received ack. %+v", c)
		// TODO: should we be calling status to confirm that it's actually true?
		a.setEvt <- *c
	case _cmdConnect:
		a.logger.Printf("received connect. %+v", c)
		a.authEvt <- *c
	default:
		a.logger.Printf("unhandled command. %+v", c)
	}
}

func socketReader(a *AsyncSocket, output chan string) error {
	scanner := bufio.NewScanner(a.conn)
	scanner.Split(doubleBraceSplit)
	for scanner.Scan() {
		output <- scanner.Text()
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	return nil
}

func keepalive(a *AsyncSocket) error {
	cmdReq := &intesishome.CommandRequest{
		Command: "get",
		Data: intesishome.CommandRequestData{
			DeviceID: a.device,
			Uid:      _keepaliveUid,
		},
	}
	if err := socketWrite(a, cmdReq); err != nil {
		a.logger.Printf("keepalive socket write error. cause: %v", err)
		return err
	}
	return nil
}

func socketWrite(a *AsyncSocket, thing interface{}) error {
	b, err := json.Marshal(thing)
	if err != nil {
		a.logger.Printf("json encode error. auth: %+v cause: %v", thing, err)
		return err
	}
	length, err := a.conn.Write(b)
	if err != nil {
		a.logger.Printf("socket write error. cause: %v", err)
		return err
	}
	if length != len(b) {
		a.logger.Printf("write length mismatch. %d != %d", length, len(b))
		return err
	}
	return nil
}

func socketRead(a *AsyncSocket) (*intesishome.CommandResponse, error) {
	cmdResp := &intesishome.CommandResponse{}
	resp := make([]byte, _readLimitBytes)

	length, err := a.conn.Read(resp)
	if err != nil {
		a.logger.Printf("socket read error. cause: %v", err)
		return nil, err
	}
	if length <= 0 {
		err := fmt.Errorf("0 byte socket read")
		a.logger.Print(err)
		return nil, err
	}
	resp = bytes.Trim(resp, "\x00") // trim any nulls from the end
	if err = json.Unmarshal(resp, &cmdResp); err != nil {
		a.logger.Printf("auth response decode error. resp: %s caused: %v", string(resp), err)
		return nil, err
	}
	return cmdResp, nil
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
		return len(data), data[:len(data)], nil
	}
	return 0, nil, nil
}
