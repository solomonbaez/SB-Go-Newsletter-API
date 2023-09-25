package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	maxEmailLength = 100
	maxNameLength  = 100
	invalidRunes   = "{}/\\<>()"
)

var (
	emailRegex = regexp.MustCompile((`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`))
)

type Subscriber struct {
	ID    string          `json:"id"`
	Email SubscriberEmail `json:"email"`
	Name  SubscriberName  `json:"name"`
}

type SubscriberEmail string
type SubscriberName string

func ParseEmail(e string) (SubscriberEmail, error) {
	// empty field check
	eTrim := strings.Trim(e, " ")
	if eTrim == "" {
		return "", errors.New("fields can not be empty or whitespace")
	}

	// length checks
	if len(e) > maxEmailLength {
		return "", fmt.Errorf("email exceeds maximum length of: %d characters", maxEmailLength)
	}

	// email format check
	if !emailRegex.MatchString(e) {
		return "", fmt.Errorf("invalid email format")
	}

	return SubscriberEmail(e), nil
}

func ParseName(n string) (SubscriberName, error) {
	// injection check
	for _, r := range n {
		c := string(r)
		if strings.Contains(invalidRunes, c) {
			return "", fmt.Errorf("invalid character in name: %v", c)
		}
	}

	// empty field check
	nTrim := strings.Trim(n, " ")
	if nTrim == "" {
		return "", errors.New("name cannot be empty or whitespace")
	}

	// length checks
	if len(n) > maxNameLength {
		return "", fmt.Errorf("name exceeds maximum length of: %d characters", maxNameLength)
	}

	return SubscriberName(n), nil
}
