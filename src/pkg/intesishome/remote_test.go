package intesishome

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	_testAuthRequest  string = `{"command":"connect_req","data":{"deviceId":0,"uid":0,"value":0,"seqNo":0,"token":12345}}`
	_testAuthResponse string = `{"command":"connect_rsp","data":{"status":"ok"}}`
	_testAuthFailure  string = `{"command":"connect_rsp","data":{"status":"err_token"}}`
	_testAuthInvalid  string = `{"command":"garbage","data":{"status":"ok"}}`
	_testSetRequest   string = `{"command":"set","data":{"deviceId":127934703953,"uid":1,"value":0,"seqNo":0,"token":0}}`
	_testSetResponse  string = `{"command":"set_ack","data":{"deviceId":127934703953,"seqNo":85,"rssi":198}}`
	_testSetInvalid   string = `{"command":"garbage","data":{"deviceId":127934703953,"seqNo":85,"rssi":198}}`
)

type mockConn struct {
	ReadFunc  ReadWriteFunc
	WriteFunc ReadWriteFunc
	PostAuth  bool
	T         *testing.T
}

type ReadWriteFunc func(*mockConn, []byte) (int, error)

func newMockConn(reader, writer ReadWriteFunc, t *testing.T) *mockConn {
	c := mockConn{
		ReadFunc:  reader,
		WriteFunc: writer,
		PostAuth:  false,
		T:         t,
	}
	return &c
}
func (c *mockConn) Read(b []byte) (n int, err error) {
	n, err = c.ReadFunc(c, b)
	return
}
func (c *mockConn) Write(b []byte) (n int, err error) {
	n, err = c.WriteFunc(c, b)
	return
}
func (c *mockConn) Close() error                       { return nil }
func (c *mockConn) LocalAddr() net.Addr                { return nil }
func (c *mockConn) RemoteAddr() net.Addr               { return nil }
func (c *mockConn) SetDeadline(t time.Time) error      { return nil }
func (c *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *mockConn) SetWriteDeadline(t time.Time) error { return nil }

type inlineControlCheck func(t *testing.T, r *ControlResponse, e error)

func TestAPICalls(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		payload string
		want    inlineControlCheck
	}{
		{
			"non 200 response",
			http.StatusBadGateway,
			testNilResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "unexpected response code")
			},
		},
		{
			"nil payload",
			http.StatusOK,
			testNilResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "unexpected response body")
			},
		},
		{
			"error payload",
			http.StatusOK,
			testErrorControlResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "unexpected response error")
			},
		},
		{
			"malformed payload",
			http.StatusOK,
			testMalformedResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.ErrorContains(t, e, "malformed payload")
			},
		},
		{
			"success payload",
			http.StatusOK,
			testValidControlResponsePayload,
			func(t *testing.T, r *ControlResponse, e error) {
				assert.NoError(t, e)
				assert.Equal(t, r.ErrorCode, 0)
				assert.Equal(t, r.Config.Inst[0].Devices[0].ID, testDeviceId)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := testControlRequest(tt.code, tt.payload)
			tt.want(t, &r, err)
		})
	}
}

func TestSocketCalls(t *testing.T) {
	t.Run("invalid endpoint", func(t *testing.T) {
		var req []byte
		ih := New("u", "p", WithTCPServer("0.0.0.0:5000"))
		_, err := socketWrite(&ih, req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "dial tcp")
	})
	t.Run("write EOF", func(t *testing.T) {
		var req []byte
		ih := New("u", "p")
		reader := func(*mockConn, []byte) (int, error) {
			return 0, nil
		}
		writer := func(*mockConn, []byte) (int, error) {
			return 0, errors.New("EOF")
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		_, err := socketWrite(&ih, req)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "socket write error")
	})
	t.Run("read EOF", func(t *testing.T) {
		ih := New("u", "p")
		reader := func(*mockConn, []byte) (int, error) {
			return 0, errors.New("EOF")
		}
		writer := func(c *mockConn, b []byte) (int, error) {
			return len(b), nil
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		_, err := socketWrite(&ih, []byte(_testAuthRequest))
		assert.Error(t, err)
		assert.ErrorContains(t, err, "socket read error")
	})
	t.Run("write mismatch", func(t *testing.T) {
		ih := New("u", "p")
		reader := func(*mockConn, []byte) (int, error) {
			return 0, nil
		}
		writer := func(c *mockConn, b []byte) (int, error) {
			return len(b) + 10, nil
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		_, err := socketWrite(&ih, []byte("string"))
		assert.Error(t, err)
		assert.ErrorContains(t, err, "write byte mismatch")
	})
}

// scenarios: bad response payloads, unexpected payloads, EOF, timeout socket close, success
func TestSet(t *testing.T) {
	t.Run("valid set response", func(t *testing.T) {
		s, err := mockHTTPServer(http.StatusOK, testValidControlResponsePayload)
		assert.NoError(t, err)
		ih := New("u", "p", WithHostname(s.URL))
		reader := func(c *mockConn, b []byte) (int, error) {
			if !c.PostAuth {
				c.PostAuth = true // auth is complete, now progress to the set
				copy(b, []byte(_testAuthResponse))
				return len([]byte(_testAuthResponse)), nil
			}
			copy(b, []byte(_testSetResponse))
			return len([]byte(_testSetResponse)), nil
		}
		writer := func(c *mockConn, b []byte) (int, error) {
			if !c.PostAuth {
				assert.Equal(c.T, _testAuthRequest, string(b))
				return len([]byte(_testAuthRequest)), nil
			}
			assert.Equal(c.T, _testSetRequest, string(b))
			return len([]byte(_testSetRequest)), nil
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		d, _ := strconv.ParseInt(testDeviceId, 0, 64)
		err = ih.Set(int64(d), 1 /* power */, 0 /* off */)
		assert.NoError(t, err)
	})
	t.Run("invalid auth token", func(t *testing.T) {
		s, err := mockHTTPServer(http.StatusOK, testValidControlResponsePayload)
		assert.NoError(t, err)
		ih := New("u", "p", WithHostname(s.URL))
		reader := func(c *mockConn, b []byte) (int, error) {
			copy(b, []byte(_testAuthFailure))
			return len([]byte(_testAuthFailure)), nil
		}
		writer := func(c *mockConn, b []byte) (int, error) {
			assert.Equal(c.T, _testAuthRequest, string(b))
			return len([]byte(_testAuthRequest)), nil
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		d, _ := strconv.ParseInt(testDeviceId, 10, 64)
		err = ih.Set(int64(d), 1 /* power */, 0 /* off */)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unexpected reply, expected: ok got: err_token")
	})
	t.Run("invalid auth response", func(t *testing.T) {
		s, err := mockHTTPServer(http.StatusOK, testValidControlResponsePayload)
		assert.NoError(t, err)
		ih := New("u", "p", WithHostname(s.URL))
		reader := func(c *mockConn, b []byte) (int, error) {
			copy(b, []byte(_testAuthInvalid))
			return len([]byte(_testAuthInvalid)), nil
		}
		writer := func(c *mockConn, b []byte) (int, error) {
			assert.Equal(c.T, _testAuthRequest, string(b))
			return len([]byte(_testAuthRequest)), nil
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		d, _ := strconv.ParseInt(testDeviceId, 10, 64)
		err = ih.Set(int64(d), 1 /* power */, 0 /* off */)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "unexpected reply, expected: connect_rsp got: garbage")
	})
	t.Run("invalid set invalid", func(t *testing.T) {
		s, err := mockHTTPServer(http.StatusOK, testValidControlResponsePayload)
		assert.NoError(t, err)
		ih := New("u", "p", WithHostname(s.URL))
		reader := func(c *mockConn, b []byte) (int, error) {
			if !c.PostAuth {
				c.PostAuth = true // auth is complete, now progress to the set
				copy(b, []byte(_testAuthResponse))
				return len([]byte(_testAuthResponse)), nil
			}
			copy(b, []byte(_testSetInvalid))
			return len([]byte(_testSetInvalid)), nil
		}
		writer := func(c *mockConn, b []byte) (int, error) {
			if !c.PostAuth {
				assert.Equal(c.T, _testAuthRequest, string(b))
				return len([]byte(_testAuthRequest)), nil
			}
			assert.Equal(c.T, _testSetRequest, string(b))
			return len([]byte(_testSetRequest)), nil
		}
		c := newMockConn(reader, writer, t)
		ih.cmdSocket = c
		d, _ := strconv.ParseInt(testDeviceId, 10, 64)
		err = ih.Set(int64(d), 1 /* power */, 0 /* off */)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "set failed, expected: set_ack got: garbage")
	})
}

func testControlRequest(responseCode int, payload string) (r ControlResponse, err error) {
	s, err := mockHTTPServer(responseCode, payload)
	if err != nil {
		fmt.Printf("mock http server problem: %v", err.Error())
		return
	}
	ih := New("u", "p", WithHostname(s.URL))
	r, err = controlRequest(&ih)
	return
}
