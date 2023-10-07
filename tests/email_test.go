package api_test

import (
	"fmt"
	"testing"

	mock "github.com/mocktools/go-smtp-mock"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

var mockClient = &clients.SMTPClient{}

func TestMockEmail_GomailClient(t *testing.T) {
	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)
	server.Start()
	port := server.PortNumber

	client := mockClient
	client.SmtpPort = port
	client.Sender = models.SubscriberEmail("user@test.com")
	fmt.Printf("%v", client)

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}
	emailContent := models.Newsletter{
		Recipient: models.SubscriberEmail("test@test.com"),
		Content:   &body,
	}

	if e := client.SendEmail(&emailContent); e != nil {
		t.Errorf("Failed to send email")
		return
	}
}
