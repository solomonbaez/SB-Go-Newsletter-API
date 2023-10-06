package api_test

import (
	"strings"
	"testing"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

func TestParseEmail(t *testing.T) {
	longEmail := "a" + strings.Repeat("a", 100) + "@test.com"

	testCases := []string{
		"", " ", longEmail, "test", "test@", "@test.com", "test.com",
	}

	for _, tc := range testCases {
		if s, e := models.ParseEmail(tc); e == nil {
			t.Fatalf(s.String())
		}
	}
}

func TestParseName(t *testing.T) {
	longName := "a" + strings.Repeat("a", 100)

	testCases := []string{
		"", " ", "a b", longName, "{", "}", "/", "\\", "<", ">", "(", ")",
	}

	for _, tc := range testCases {
		if s, e := models.ParseName(tc); e == nil {
			t.Fatalf(s.String())
		}
	}
}
