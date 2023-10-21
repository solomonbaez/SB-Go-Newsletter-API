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
)

var (
	emailRegex = regexp.MustCompile((`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`))
)

type Subscriber struct {
	ID     string          `json:"id"`
	Email  SubscriberEmail `json:"email" binding:"required"`
	Name   SubscriberName  `json:"name" binding:"required"`
	Status string          `json:"status"`
}

type SubscriberEmail string

func (email SubscriberEmail) String() string {
	return string(email)
}

func (email SubscriberEmail) MarshalJSON() ([]byte, error) {
	return json.Marshal(email.String())
}

func (email *SubscriberEmail) UnmarshalJSON(data []byte) (err error) {
	var emailData string
	if e := json.Unmarshal(data, &emailData); e != nil {
		err = fmt.Errorf("failed to unmarshal subscriber email JSON: %w", e)
		return
	}

	*email = SubscriberEmail(emailData)
	return
}

func ParseEmail(email string) (subscriber SubscriberEmail, err error) {
	// empty field check
	emptyField := strings.Trim(email, " ")
	if emptyField == "" {
		err = errors.New("fields can not be empty or whitespace")
		return
	}

	// length checks
	if len(email) > maxEmailLength {
		err = fmt.Errorf("email exceeds maximum length of: %d characters", maxEmailLength)
		return
	}

	// email format check
	if !emailRegex.MatchString(email) {
		err = fmt.Errorf("invalid email format")
		return
	}

	subscriber = SubscriberEmail(email)
	return
}

type SubscriberName string

func (name SubscriberName) String() string {
	return string(name)
}

func (name SubscriberName) MarshalJSON() ([]byte, error) {
	return json.Marshal(name.String())
}

func (name *SubscriberName) UnmarshalJSON(data []byte) (err error) {
	var nameData string
	if e := json.Unmarshal(data, &nameData); e != nil {
		err = fmt.Errorf("failed to unmarshal subscriber name JSON: %w", e)
		return
	}

	*name = SubscriberName(nameData)
	return
}

func ParseName(name string) (subscriber SubscriberName, err error) {
	// injection check
	for _, r := range name {
		c := string(r)
		if strings.Contains(InvalidRunes, c) {
			err = fmt.Errorf("invalid character in name: %v", c)
			return
		}
	}

	// empty field check
	emptyField := strings.Trim(name, " ")
	if emptyField == "" {
		err = errors.New("name cannot be empty or whitespace")
		return
	}

	// length checks
	if len(name) > maxNameLength {
		err = fmt.Errorf("name exceeds maximum length of: %d characters", maxNameLength)
		return
	}

	subscriber = SubscriberName(name)
	return
}
