package mock

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/nullify005/service-intesis/pkg/intesishome"
)

const (
	DefaultTCPListen      string        = "127.0.0.1:5000"
	DefaultReadLimitBytes int           = 4096
	DefaultReadTimeout    time.Duration = 30 * time.Second
	rssiCeil              int           = 256
)

// wrapper around the net.Conn which we can operate on within a thread
type Conn struct {
	t              uuid.UUID
	c              net.Conn
	readLimitBytes int
	authenticated  bool
}

type TCPServer struct {
	Listen         string
	ReadLimitBytes int
	ReadTimeout    time.Duration
}

type Option func(t *TCPServer)

// builds & returns a TCPServer
func NewTCPServer(opts ...Option) TCPServer {
	t := TCPServer{
		Listen:         DefaultTCPListen,
		ReadLimitBytes: DefaultReadLimitBytes,
		ReadTimeout:    DefaultReadTimeout,
	}
	for _, o := range opts {
		o(&t)
	}
	return t
}

// sets an alternate listen:port
func WithTCPListen(l string) Option {
	return func(t *TCPServer) {
		t.Listen = l
	}
}

// sets the socket read byte limit
func WithTCPReadLimitBytes(b int) Option {
	return func(t *TCPServer) {
		t.ReadLimitBytes = b
	}
}

// sets the socket read timeout
func WithTCPReadTimeout(d time.Duration) Option {
	return func(t *TCPServer) {
		t.ReadTimeout = d
	}
}

// run the TCPServer listener responding to incoming requests
func (t *TCPServer) Run() error {
	l, err := net.Listen("tcp4", t.Listen)
	if err != nil {
		return err
	}
	defer l.Close()
	log.Printf("TCPServer listening on: %s", t.Listen)
	for {
		var c Conn
		conn, err := l.Accept()
		if err != nil {
			log.Print(err.Error())
			continue
		}
		c.c = conn
		c.t = uuid.New()
		c.readLimitBytes = t.ReadLimitBytes
		c.c.SetDeadline(time.Now().Add(t.ReadTimeout))
		go handleConn(&c)
	}
}

// return a byte array from a Scanner.Scan which is delimited by }}
func splitDoubleEndBrace(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'}', '}'}); i >= 0 {
		// We have a full }} terminated line.
		return i + 2, data[0 : i+2], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}

// TODO: move this into the scanner func?
func jsonPayload(p string) (string, error) {
	b := []byte(p)
	for i := 0; i < len(p); i++ {
		if b[i] == '{' {
			return string(b[i:len(p)]), nil
		}
	}
	return "", fmt.Errorf("unable to locate start delimiter")
}

// TODO: compress this
func handleConn(c *Conn) {
	defer c.c.Close()
	limit := &io.LimitedReader{R: c.c, N: int64(c.readLimitBytes)}
	log.Printf("(%s) received connection from: %s", c.t, c.c.RemoteAddr().String())
	s := bufio.NewScanner(limit)
	s.Split(splitDoubleEndBrace)
	for s.Scan() {
		j, err := jsonPayload(s.Text())
		if err != nil {
			log.Printf("(%s) ignoring invalid payload: %s", c.t, s.Text())
			continue
		}
		go handlePayload(c, j)
	}
	if err := s.Err(); err != nil {
		log.Printf("(%s) %s", c.t, err.Error())
	}
	log.Printf("(%s) closed connection from: %s", c.t, c.c.RemoteAddr().String())
}

/*
	writeCommand  string = `{"command":"set","data":{"deviceId":%v,"uid":%v,"value":%v,"seqNo":0}}`
	writeAuth     string = `{"command":"connect_req","data":{"token":%v}}`
		cmd.Command = "connect_req"
	commandConnOk string = "connect_rsp"
		cmd.Command = "set"
	commandSetAck string = "set_ack"
		cmdResp.Data.Status != commandAuthOk
	commandAuthOk string = "ok"

	{"command":"connect_req","data":{"deviceId":12345,"token":12345}}
	{"command":"connect_req","data":{"deviceId":12345}}

	{"command":"set","data":{"deviceId":12345,"uid":1,"value":1,"seqNo":0}}

    async def _parse_response(self, decoded_data):
        _LOGGER.debug("%s API Received: %s", self._device_type, decoded_data)
        resp = json.loads(decoded_data)
        # Parse response
        if resp["command"] == "connect_rsp":
            # New connection success
            if resp["data"]["status"] == "ok":
                _LOGGER.info("%s successfully authenticated", self._device_type)
                self._connected = True
                self._connecting = False
                self._connection_retries = 0
                await self._send_update_callback()
        elif resp["command"] == "status":
            # Value has changed
            self._update_device_state(
                resp["data"]["deviceId"],
                resp["data"]["uid"],
                resp["data"]["value"],
            )
            if resp["data"]["uid"] != 60002:
                await self._send_update_callback(
                    device_id=str(resp["data"]["deviceId"])
                )
        elif resp["command"] == "rssi":
            # Wireless strength has changed
            self._update_rssi(resp["data"]["deviceId"], resp["data"]["value"])
        return

	received response: {"command":"connect_rsp","data":{"status":"ok"}}
	received response: {"command":"set_ack","data":{"deviceId":127934703953,"seqNo":85,"rssi":198}}

DEBUG|socketWrite| sending request: {"command":"set","data":{"deviceId":127934703953,"uid":-1,"value":0,"seqNo":0,"token":0}}
DEBUG|socketWrite| received response: {"command":"set_ack","data":{"deviceId":127934703953,"seqNo":0,"rssi":198}}

DEBUG|controlRequest| token: 575497412 server: 212.36.84.207 5210
DEBUG|socketWrite| sending request: {"command":"connect_req","data":{"deviceId":0,"uid":0,"value":0,"seqNo":0,"token":575497412}}
DEBUG|socketWrite| received response: {"command":"connect_rsp","data":{"status":"ok"}}
DEBUG|socketWrite| sending request: {"command":"set","data":{"deviceId":127934703953,"uid":1,"value":0,"seqNo":0,"token":0}}
failure to read from socket with: EOF
exit status 1

we don't seem to error at all on bad input, we just ack & then do nothing

	ERRORS:
		bad token
			{"command":"connect_rsp","data":{"status":"err_token"}}
*/

// TODO: add in more rigourous request checking
// TODO: write tests
func handlePayload(c *Conn, p string) {
	rand.Seed(time.Now().UnixNano()) // seed for the RSSI responses
	var request intesishome.CommandRequest
	var response intesishome.CommandResponse
	if err := json.Unmarshal([]byte(p), &request); err != nil {
		log.Printf("(%s) cannot unmarshal: %s", c.t, err.Error())
		return
	}
	log.Printf("(%s) recieved payload: %#v", c.t, request)
	switch request.Command {
	case "connect_req":
		response.Command = "connect_rsp"
		response.Data.Status = "ok"
		if request.Data.Token == 0 {
			// There is no token, so we give back an error
			response.Data.Status = "err_token"
		}
		r, err := json.Marshal(response)
		if err != nil {
			log.Printf("(%s) cannot marshal response: %s", c.t, err.Error())
			return
		}
		c.authenticated = true
		c.c.Write(r)
	case "set":
		if !c.authenticated {
			// have to be authed first, behaviour is that we EOF
			c.c.Close()
			return
		}
		// TODO: what does the API respond to when you don't set various params ... aparently nothing
		// TODO: when we didn't auth? we just EOF
		response.Command = "set_ack"
		response.Data.Rssi = rand.Intn(rssiCeil)
		if request.Data.DeviceID == 0 {
			// the real thing doesn't care
			log.Printf("(%s) ignoring attempt to set on a missing device", c.t)
			return
		}
		if request.Data.Uid == 0 {
			// the real thing doesn't care
			log.Printf("(%s) ignoring attempt to set on an empty uid", c.t)
			return
		}
		response.Data.DeviceID = request.Data.DeviceID
		r, err := json.Marshal(response)
		if err != nil {
			log.Printf("(%s) cannot marshal response: %s", c.t, err.Error())
			return
		}
		c.c.Write(r)
	default:
		log.Printf("(%s) ignoring invalid payload, command was empty: %s", c.t, p)
		return
	}
}
