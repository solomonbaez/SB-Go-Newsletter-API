package api_test

import (
	"testing"
	"time"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
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

	body := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "testing",
	}
	emailContent := models.Newsletter{
		Recipient: models.SubscriberEmail("test@test.com"),
		Content:   &body,
	}
	if e := client.SendEmail(&emailContent); e != nil {
		t.Errorf("Failed to send email: %v", e)
		return
	}

	// asynchronous testing
	done := make(chan struct{})
	go func() {
		for {
			recievedEmails := server.GetEmails()
			t.Logf("Recieved emails: %v, Content: %v", recievedEmails, emailContent)

			if len(recievedEmails) == 1 {
				recieved := recievedEmails[0]
				// dereference newsletter.Content pointers
				if recieved.Recipient == emailContent.Recipient && *recieved.Content == *emailContent.Content {
					close(done)
					return
				}
				t.Errorf("Expected %v, got %v", emailContent, recieved)
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("Timed out waiting for email processing")
	}

}
