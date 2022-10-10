package secrets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	goodFile string = "assets/good.yaml"
	badFile  string = "assets/bad.yaml"
	jsonFile string = "assets/malformed.yaml"
)

func TestGoodYaml(t *testing.T) {
	s, err := Read(goodFile)
	assert.NoError(t, err)
	assert.Equal(t, s.Username, "a")
	assert.Equal(t, s.Password, "b")
}

func TestBadYaml(t *testing.T) {
	_, err := Read(badFile)
	assert.Error(t, err)
}

func TestBadFile(t *testing.T) {
	_, err := Read(jsonFile)
	assert.Error(t, err)
}
