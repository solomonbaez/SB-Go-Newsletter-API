package models

import (
	"fmt"

	"reflect"
)

type Newsletter struct {
	Recipient SubscriberEmail
	Content   *Body
}

type Body struct {
	Title string `json:"title" binding:"required"`
	Text  string `json:"text" binding:"required"`
	Html  string `json:"html" binding:"required"`
}

func ParseNewsletter(newsletter interface{}) (err error) {
	value := reflect.ValueOf(newsletter).Elem()
	nFields := value.NumField()

	for i := 0; i < nFields; i++ {
		field := value.Field(i)
		if !field.IsValid() || reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			name := value.Type().Field(i).Name
			err = fmt.Errorf("field: %s cannot be empty", name)
			break
		}
	}

	return
}
