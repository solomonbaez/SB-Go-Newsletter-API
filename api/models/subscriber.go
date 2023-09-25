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
	ID    string `json:"id"`
	Email SubscriberEmail
	Name  SubscriberName
}

type SubscriberEmail struct {
	Email string `json:"email"`
}
type SubscriberName struct {
	Name string `json:"name"`
}

func ParseEmail(e string) SubscriberEmail {
	// empty field check
	eTrim := strings.Trim(e, " ")
	if eTrim == "" {
		panic(errors.New("fields can not be empty or whitespace"))
	}

	// length checks
	if len(e) > maxEmailLength {
		panic(fmt.Errorf("email exceeds maximum length of: %d characters", maxEmailLength))
	}

	// email format check
	if !emailRegex.MatchString(e) {
		panic(fmt.Errorf("invalid email format"))
	}

	email := SubscriberEmail{
		Email: e,
	}

	return email
}

func ParseName(n string) SubscriberName {
	// injection check
	for _, r := range n {
		c := string(r)
		if strings.Contains(invalidRunes, c) {
			panic(fmt.Errorf("invalid character in name: %v", c))
		}
	}

	// empty field check
	nTrim := strings.Trim(n, " ")
	if nTrim == "" {
		panic(errors.New("name cannot be empty or whitespace"))
	}

	// length checks
	if len(n) > maxNameLength {
		panic(fmt.Errorf("name exceeds maximum length of: %d characters", maxNameLength))
	}

	name := SubscriberName{
		Name: n,
	}

	return name
}
