package api_test

import (
	"fmt"
	"testing"

	mock "github.com/mocktools/go-smtp-mock"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/clients"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

var mockClient = &clients.SMTPClient{}

func TestMockEmail_ValidEmail_Passes(t *testing.T) {
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

func TestMockEmail_NoRecipient_Fails(t *testing.T) {
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
		Content: &body,
	}

	if e := client.SendEmail(&emailContent); e == nil {
		t.Errorf("Failed to invalid filter email")
		return
	}
}

func TestMockEmail_InvalidRecipient_Fails(t *testing.T) {
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
		Recipient: models.SubscriberEmail("test.com"),
		Content:   &body,
	}

	if e := client.SendEmail(&emailContent); e == nil {
		t.Errorf("Failed to filter invalid email")
		return
	}
}
