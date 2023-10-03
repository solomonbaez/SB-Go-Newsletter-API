package api_test

import (
	"testing"

	api "github.com/solomonbaez/SB-Go-Newsletter-API/tests"
)

func TestMockEmail(t *testing.T) {
	server := api.NewMockSMTPServer()
	server.Start()
	defer server.Stop()

	addr := server.GetAddr()
	t.Logf("Server address is: %s", addr)

	client := api.NewMockSMTPClient(addr)
	t.Logf("Client connected to: %s", client.Addr)

	emailContent := api.MockEmail{
		Title: "testing",
		Text:  "testing",
		Html:  "testing",
	}
	if e := client.SendEmail(emailContent); e != nil {
		t.Errorf("Failed to send email: %v", e)
	}

	recievedEmails := server.GetEmails()
	t.Logf("Recieved emails: %v", recievedEmails)

	if len(recievedEmails) != 1 {
		t.Errorf("Expected 1 recieved email, got %d", len(recievedEmails))
	}

	if recievedEmails[0] != emailContent {
		t.Errorf("Expected %v, got %v", recievedEmails[0], emailContent)
	}
}
