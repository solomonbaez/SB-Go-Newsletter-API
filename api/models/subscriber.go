package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	maxEmailLength = 100
	maxNameLength  = 100
	invalidRunes   = "{}/\\<>() "
)

var (
	emailRegex = regexp.MustCompile((`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`))
)

type Subscriber struct {
	ID     string          `json:"id"`
	Email  SubscriberEmail `json:"email"`
	Name   SubscriberName  `json:"name"`
	Status string          `json:"status"`
}

type SubscriberEmail string

func (email SubscriberEmail) String() string {
	return string(email)
}

func (email SubscriberEmail) MarshalJSON() ([]byte, error) {
	return json.Marshal(email.String())
}

func (email *SubscriberEmail) UnmarshalJSON(data []byte) error {
	var emailData string
	if e := json.Unmarshal(data, &emailData); e != nil {
		return e
	}

	*email = SubscriberEmail(emailData)
	return nil
}

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

type SubscriberName string

func (name SubscriberName) String() string {
	return string(name)
}

func (name SubscriberName) MarshalJSON() ([]byte, error) {
	return json.Marshal(name.String())
}

func (name *SubscriberName) UnmarshalJSON(data []byte) error {
	var nameData string
	if e := json.Unmarshal(data, &nameData); e != nil {
		return e
	}

	*name = SubscriberName(nameData)
	return nil
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
