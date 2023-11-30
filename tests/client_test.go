package api_test

import (
	"fmt"
	"testing"

	mock "github.com/mocktools/go-smtp-mock"
	"github.com/solomonbaez/hyacinth/api/clients"
	"github.com/solomonbaez/hyacinth/api/models"
)

var mockClient = &clients.SMTPClient{}

func TestMockEmail_ValidEmail_Passes(t *testing.T) {
	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)
	server.Start()
	port := server.PortNumber
	defer server.Stop()

	client := mockClient
	client.SmtpPort = port

	sender := models.SubscriberEmail("user@example.com")
	client.Sender = &sender

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}

	recipient := models.SubscriberEmail("test@example.com")
	emailContent := models.Newsletter{
		Recipient: recipient,
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
	defer server.Stop()

	client := mockClient
	client.SmtpPort = port

	sender := models.SubscriberEmail("user@example.com")
	client.Sender = &sender

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}
	emailContent := models.Newsletter{
		Content: &body,
	}

	if e := client.SendEmail(&emailContent); e == nil {
		t.Errorf("Failed to filter invalid email")
		return
	}
}

func TestMockEmail_InvalidRecipient_Fails(t *testing.T) {
	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)
	server.Start()
	port := server.PortNumber
	defer server.Stop()

	client := mockClient
	client.SmtpPort = port

	sender := models.SubscriberEmail("user@example.com")
	client.Sender = &sender

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}

	recipient := models.SubscriberEmail("example.com")
	emailContent := models.Newsletter{
		Recipient: recipient,
		Content:   &body,
	}

	if e := client.SendEmail(&emailContent); e == nil {
		t.Errorf("Failed to filter invalid email")
		return
	}
}

func TestMockEmail_InvalidBody_Fails(t *testing.T) {
	cfg := mock.ConfigurationAttr{}
	server := mock.New(cfg)
	server.Start()
	port := server.PortNumber
	defer server.Stop()

	client := mockClient
	client.SmtpPort = port
	sender := models.SubscriberEmail("user@example.com")
	client.Sender = &sender
	fmt.Printf("%v", client)

	testCases := []models.Body{
		{
			Title: "",
			Text:  "testing",
			Html:  "<p>testing</p>",
		},
		{
			Title: "testing",
			Text:  "",
			Html:  "<p>testing</p>",
		},
		{
			Title: "testing",
			Text:  "testing",
			Html:  "",
		},
	}

	recipient := models.SubscriberEmail("test@example.com")
	for _, tc := range testCases {
		emailContent := models.Newsletter{
			Recipient: recipient,
			Content:   &tc,
		}

		if e := client.SendEmail(&emailContent); e == nil {
			t.Errorf("Failed to filter invalid email")
			return
		}
	}
}
