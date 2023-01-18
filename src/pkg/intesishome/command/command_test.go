package command

import (
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	_testDeviceId     int64  = 127934703953
	_testValidToken   int    = 12345
	_testAuthRequest  string = `{"command":"connect_req","data":{"deviceId":0,"uid":0,"value":0,"seqNo":0,"token":12345}}`
	_testAuthResponse string = `{"command":"connect_rsp","data":{"status":"ok"}}`
	_testAuthFailure  string = `{"command":"connect_rsp","data":{"status":"err_token"}}`
	_testAuthInvalid  string = `{"command":"garbage","data":{"status":"ok"}}`
	_testSetRequest   string = `{"command":"set","data":{"deviceId":127934703953,"uid":1,"value":0,"seqNo":0,"token":0}}`
	_testSetResponse  string = `{"command":"set_ack","data":{"deviceId":127934703953,"seqNo":85,"rssi":198}}`
	_testSetInvalid   string = `{"command":"garbage","data":{"deviceId":127934703953,"seqNo":85,"rssi":198}}`
)

type mockConn struct {
	receiveFunc ReadWriteFunc
	respondFunc ReadWriteFunc
	respondWith chan []byte
	t           *testing.T
}

type ReadWriteFunc func(*mockConn, []byte) (int, error)

func newMockConn(responder, receiver ReadWriteFunc, t *testing.T) *mockConn {
	c := mockConn{
		receiveFunc: receiver,
		respondFunc: responder,
		respondWith: make(chan []byte),
		t:           t,
	}
	return &c
}
func (c *mockConn) Read(b []byte) (n int, err error) {
	n, err = c.respondFunc(c, b)
	return
}
func (c *mockConn) Write(b []byte) (n int, err error) {
	n, err = c.receiveFunc(c, b)
	return
}
func (c *mockConn) Close() error                       { return nil }
func (c *mockConn) LocalAddr() net.Addr                { return nil }
func (c *mockConn) RemoteAddr() net.Addr               { return nil }
func (c *mockConn) SetDeadline(t time.Time) error      { return nil }
func (c *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mockConn) SetWriteDeadline(t time.Time) error { return nil }

func TestSocketCalls(t *testing.T) {
	t.Run("invalid endpoint", func(t *testing.T) {
		c := New()
		err := c.Connect("")
		assert.Error(t, err)
		assert.ErrorContains(t, err, "dial tcp: missing address")
	})
	t.Run("write EOF", func(t *testing.T) {
		var req []byte
		reader := func(*mockConn, []byte) (int, error) {
			return 0, nil
		}
		writer := func(*mockConn, []byte) (int, error) {
			return 0, errors.New("EOF")
		}
		mock := newMockConn(reader, writer, t)
		c := New(WithConn(mock))
		err := socketWrite(c, req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "socket write error")
	})
	t.Run("read EOF", func(t *testing.T) {
		responder := func(*mockConn, []byte) (int, error) {
			return 0, errors.New("EOF")
		}
		receiver := func(c *mockConn, b []byte) (int, error) {
			return len(b), nil
		}
		outputC := make(chan string)
		errC := make(chan error)
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		go socketReader(c, outputC, errC)
		err := <-errC
		assert.Error(t, err)
		assert.ErrorContains(t, err, "EOF")
	})
	t.Run("write mismatch", func(t *testing.T) {
		responder := func(*mockConn, []byte) (int, error) {
			return 0, nil
		}
		receiver := func(c *mockConn, b []byte) (int, error) {
			return len(b) + 10, nil
		}
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		err := socketWrite(c, []byte("string"))
		assert.Error(t, err)
		assert.ErrorContains(t, err, "write length mismatch")
	})
}

// scenarios: bad response payloads, unexpected payloads, EOF, timeout socket close, success
func TestSetHandler(t *testing.T) {
	t.Run("set response success", func(t *testing.T) {
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		c.Token(_testValidToken)
		go c.Listen()
		err := c.Set(_testDeviceId, 1 /* power */, 0 /* off */)
		assert.NoError(t, err)
	})
	t.Run("set response timeout", func(t *testing.T) {
		receiver := func(c *mockConn, b []byte) (int, error) {
			switch string(b) {
			case _testAuthRequest:
				c.respondWith <- []byte(_testAuthResponse)
			case _testSetRequest:
				time.Sleep(10 * time.Second) // block without sending a response
			}
			return len(b), nil
		}
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		c.Token(_testValidToken)
		go c.Listen()
		err := c.Set(_testDeviceId, 1 /* power */, 0 /* off */)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "timeout waiting for set response")
	})
	t.Run("invalid auth token", func(t *testing.T) {
		receiver := func(c *mockConn, b []byte) (int, error) {
			switch string(b) {
			case _testAuthRequest:
				c.respondWith <- []byte(_testAuthFailure)
			}
			return len(b), nil
		}
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		c.Token(_testValidToken)
		go c.Listen()
		time.Sleep(3 * time.Second) /* wait for the auth to fail */
		err := c.Err()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid auth response")
		// assert.ErrorContains(t, err, "got: err_token")
	})
	t.Run("invalid auth token", func(t *testing.T) {
		receiver := func(c *mockConn, b []byte) (int, error) {
			switch string(b) {
			case _testAuthRequest:
				c.respondWith <- []byte(_testAuthInvalid)
			}
			return len(b), nil
		}
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		c.Token(_testValidToken)
		go c.Listen()
		time.Sleep(3 * time.Second) /* wait for the auth to fail */
		err := c.Err()
		assert.Error(t, err)
		assert.ErrorContains(t, err, "invalid auth response")
		// assert.ErrorContains(t, err, "got: garbage")
	})
	t.Run("set response invalid", func(t *testing.T) {
		receiver := func(c *mockConn, b []byte) (int, error) {
			switch string(b) {
			case _testAuthRequest:
				c.respondWith <- []byte(_testAuthResponse)
			case _testSetRequest:
				c.respondWith <- []byte(_testSetInvalid)
			}
			return len(b), nil
		}
		mock := newMockConn(responder, receiver, t)
		c := New(WithConn(mock))
		c.Token(_testValidToken)
		go c.Listen()
		err := c.Set(_testDeviceId, 1 /* power */, 0 /* off */)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "timeout waiting for set response")
	})
}

func seqNo() int {
	return int(time.Now().Unix())
}

func receiver(c *mockConn, b []byte) (int, error) {
	c.t.Logf("received message: %s", b)
	j := &Request{}
	r := &Response{}
	err := json.Unmarshal(b, j)
	if err != nil {
		return len(b), err
	}
	r.Data.SeqNo = seqNo()
	r.Data.DeviceID = j.Data.DeviceID
	switch j.Command {
	case _cmdConnect:
		r.Command = _cmdConnectAck
		r.Data.Status = _cmdConnectOk
	case _cmdKeepalive:
		r.Command = _cmdStatus
	case _cmdSet:
		r.Command = _cmdSetAck
	}
	rBytes, err := json.Marshal(r)
	if err != nil {
		return len(b), err
	}
	c.respondWith <- rBytes
	return len(b), nil
}

func responder(c *mockConn, b []byte) (int, error) {
	response := <-c.respondWith
	c.t.Logf("responding with: %s", response)
	copy(b, response)
	return len(b), nil
}
