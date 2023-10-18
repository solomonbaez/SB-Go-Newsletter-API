package models

import (
	"fmt"

	"reflect"
)

type Newsletter struct {
	Recipient SubscriberEmail
	Content   *Body
	// Key       string
}

type Body struct {
	Title string `json:"title"`
	Text  string `json:"text"`
	Html  string `json:"html"`
}

func ParseNewsletter(newsletter interface{}) error {
	v := reflect.ValueOf(newsletter).Elem()
	nFields := v.NumField()

	for i := 0; i < nFields; i++ {
		field := v.Field(i)
		valid := field.IsValid() && !field.IsZero()
		if !valid {
			name := v.Type().Field(i).Name
			return fmt.Errorf("field: %s cannot be empty", name)
		}
	}

	return nil
}
