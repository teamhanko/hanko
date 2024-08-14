package config

import (
	"errors"
	"strings"
)

type Service struct {
	// `name` determines the name of the service.
	// This value is used, e.g. in the subject header of outgoing emails.
	Name string `yaml:"name" json:"name,omitempty" koanf:"name"`
}

func (s *Service) Validate() error {
	if len(strings.TrimSpace(s.Name)) == 0 {
		return errors.New("field name must not be empty")
	}
	return nil
}
