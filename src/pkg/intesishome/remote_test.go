package intesishome

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
