package secrets

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Secrets struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func Read(path string) (s *Secrets, err error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return s, err
	}
	err = yaml.Unmarshal(body, &s)
	if err != nil {
		return s, err
	}
	if s.Username == "" || s.Password == "" {
		err = fmt.Errorf("username or password fields cannot be empty! %#v", s)
		return s, err
	}
	return
}
