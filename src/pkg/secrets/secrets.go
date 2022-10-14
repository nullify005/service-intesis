package secrets

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Secrets struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func Read(path string) (s *Secrets, err error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return
	}
	d := yaml.NewDecoder(strings.NewReader(string(body)))
	d.KnownFields(true)
	err = d.Decode(&s)
	return
}
