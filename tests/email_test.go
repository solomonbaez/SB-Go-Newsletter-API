package api_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
	api "github.com/solomonbaez/SB-Go-Newsletter-API/test_helpers"
	helpers "github.com/solomonbaez/SB-Go-Newsletter-API/test_helpers"
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

func TestMockEmail_EmptyFields_Fails(t *testing.T) {
	server := api.NewMockSMTPServer()
	server.Start()
	defer server.Stop()

	addr := server.GetAddr()
	t.Logf("Server address is: %s", addr)

	client := api.NewMockSMTPClient(addr)
	t.Logf("Client connected to: %s", client.Addr)

	var testCases []*models.Body
	testBody := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}

	var b models.Body
	for i := 0; i < 3; i++ {
		b = testBody
		if i == 0 {
			b.Title = ""
		} else if i == 1 {
			b.Text = ""
		} else {
			b.Html = ""
		}

		testCases = append(testCases, &b)
	}

	for _, tc := range testCases {
		fmt.Printf("tc: %v", tc)
		emailContent := models.Newsletter{
			Recipient: models.SubscriberEmail("test@test.com"),
			Content:   tc,
		}
		if e := client.SendEmail(&emailContent); e == nil {
			t.Errorf("Failed to filter email: %v", e)
		}
	}
}

func TestMockEmail_EmptyRecipient_Fails(t *testing.T) {
	server := api.NewMockSMTPServer()
	server.Start()
	defer server.Stop()

	addr := server.GetAddr()
	t.Logf("Server address is: %s", addr)

	client := api.NewMockSMTPClient(addr)
	t.Logf("Client connected to: %s", client.Addr)

	testBody := models.Body{
		Title: "testing",
		Text:  "testing",
		Html:  "<p>testing</p>",
	}
	emailContent := models.Newsletter{
		Recipient: models.SubscriberEmail(""),
		Content:   &testBody,
	}
	if e := client.SendEmail(&emailContent); e == nil {
		t.Errorf("Failed to filter email: %v", e)
	}
}

func TestMockEmail_GomailClient(t *testing.T) {
	server := api.NewMockSMTPServer()
	server.Start()
	defer server.Stop()

	addr := server.GetAddr()
	tmp := strings.Split(addr, ":")
	port, e := strconv.Atoi(tmp[len(tmp)-1])
	if e != nil {
		t.Errorf("Failed to parse port")
		return
	}

	client := helpers.TestClient
	client.SmtpServer = "[::]"
	client.SmtpPort = port
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
