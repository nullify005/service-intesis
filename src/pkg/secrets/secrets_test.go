package secrets

import (
	"testing"
)

const (
	goodFile string = "assets/good.yaml"
	badFile  string = "assets/bad.yaml"
	jsonFile string = "assets/malformed.yaml"
)

func TestGoodYaml(t *testing.T) {
	s, err := Read(goodFile)
	if err != nil {
		t.Errorf("expecting a nil error response but got: %v", err)
		return
	}
	if s.Username != "a" || s.Password != "b" {
		t.Errorf("expected valid username & password but got: %#v", s)
	}
}

func TestBadYaml(t *testing.T) {
	_, err := Read(badFile)
	if err == nil {
		t.Errorf("expecting an error response!")
	}
}

func TestBadFile(t *testing.T) {
	_, err := Read(jsonFile)
	if err == nil {
		t.Errorf("expecting an error response!")
	}
}
