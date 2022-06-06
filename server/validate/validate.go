package validate

import (
	"errors"
	"regexp"
)

func Username(username string) error {
	return text(username, "username")
}

func Password(password string) error {
	return text(password, "password")
}

func text(text, descriptor string) error {
	if len(text) < 1 || len(text) > 127 {
		return errors.New(descriptor + " has invalid length")
	}

	if matched, _ := regexp.MatchString(`^[_\-\.0-9a-z]+$`, text); !matched {
		return errors.New(descriptor + " has invalid syntax")
	}

	return nil
}
