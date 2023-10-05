package api

import (
	"bufio"
	"fmt"
	"net"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/handlers"
	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

type MockSMTPClient struct {
	Addr string
}

func NewMockSMTPClient(addr string) *MockSMTPClient {
	return &MockSMTPClient{
		Addr: addr,
	}
}

func (c *MockSMTPClient) SendEmail(email *models.Newsletter) error {
	var e error

	if e := handlers.ParseNewsletter(email); e != nil {
		return e
	}
	if e := handlers.ParseNewsletter(email.Content); e != nil {
		return e
	}

	conn, e := net.Dial("tcp", c.Addr)
	if e != nil {
		return e
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	_, e = fmt.Fprintf(writer, "Recipient: %s\n", email.Recipient)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(writer, "Title: %s\n", email.Content.Title)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(writer, "Text: %s\n", email.Content.Text)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(writer, "Html: %s\n", email.Content.Html)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(writer, ".\n")
	if e != nil {
		return e
	}

	e = writer.Flush()
	if e != nil {
		return e
	}

	return nil
}
