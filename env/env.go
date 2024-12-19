package env

import (
	"fmt"
	"os"
)

type str struct {
	key      string
	required bool
	target   *string
}

func String(target *string, key string) *str {
	return &str{
		key:      key,
		required: false,
		target:   target,
	}
}

func (s *str) Required() *str {
	s.required = true
	return s
}

func (s *str) Resolve() error {
	v := os.Getenv(s.key)
	if s.required && v == "" {
		return fmt.Errorf("env '%s' required but empty", s.key)
	}

	*s.target = v
	return nil
}
