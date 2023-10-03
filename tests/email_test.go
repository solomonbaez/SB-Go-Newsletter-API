package api_test

import (
	"testing"
	"time"

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

	// asynchronous testing
	done := make(chan struct{})
	go func() {
		for {
			recievedEmails := server.GetEmails()
			t.Logf("Recieved emails: %v, Content: %v", recievedEmails, emailContent)

			if len(recievedEmails) == 1 {
				if recievedEmails[0] == emailContent {
					close(done)
					return
				}
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
