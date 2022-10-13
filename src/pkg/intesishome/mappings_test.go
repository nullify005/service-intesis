package intesishome

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeState(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value int
		want  interface{}
	}{
		{
			"valid response",
			"power",
			1,
			"on",
		},
		{
			"invalid key",
			"unknown",
			65535,
			65535,
		},
		{
			"invalid value",
			"power",
			-1,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mVal := DecodeState(tt.key, tt.value)
			assert.Equal(t, tt.want, mVal)
		})
	}
}

func TestDecodeUid(t *testing.T) {
	tests := []struct {
		name string
		uid  int
		want interface{}
	}{
		{
			"valid uid",
			1,
			"power",
		},
		{
			"invalid uid",
			65535,
			"65535",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := DecodeUid(tt.uid)
			assert.Equal(t, v, tt.want)
		})
	}

}
