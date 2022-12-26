package cloud

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDeviceId                    string = "127934703953"
	testNilResponsePayload          string = "../assets/tests/nilResponse.json"
	testValidControlResponsePayload string = "../assets/tests/validControlResponse.json"
	testErrorControlResponsePayload string = "../assets/tests/noMethodResponse.json"
	testMalformedResponsePayload    string = "../assets/tests/malformedControlResponse.txt"
)

type inlineControlCheck func(t *testing.T, r *Response, e error)

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
			func(t *testing.T, r *Response, e error) {
				assert.ErrorContains(t, e, "unexpected response code")
			},
		},
		{
			"nil payload",
			http.StatusOK,
			testNilResponsePayload,
			func(t *testing.T, r *Response, e error) {
				assert.ErrorContains(t, e, "unexpected response body")
			},
		},
		{
			"error payload",
			http.StatusOK,
			testErrorControlResponsePayload,
			func(t *testing.T, r *Response, e error) {
				assert.ErrorContains(t, e, "unexpected response error")
			},
		},
		{
			"malformed payload",
			http.StatusOK,
			testMalformedResponsePayload,
			func(t *testing.T, r *Response, e error) {
				assert.ErrorContains(t, e, "malformed payload")
			},
		},
		{
			"success payload",
			http.StatusOK,
			testValidControlResponsePayload,
			func(t *testing.T, r *Response, e error) {
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

func testControlRequest(responseCode int, payload string) (r Response, err error) {
	s, err := mockHTTPServer(responseCode, payload)
	if err != nil {
		fmt.Printf("mock http server problem: %v", err.Error())
		return
	}
	ih := New("u", "p", WithHostname(s.URL))
	r, err = status(ih)
	return
}

func mockHTTPServer(responseCode int, payloadFile string) (s *httptest.Server, err error) {
	body, err := os.ReadFile(payloadFile)
	if err != nil {
		return
	}
	s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(responseCode)
		w.Write(body)
	}))
	return
}
