package api

import (
	"bufio"
	"fmt"
	"net"
)

type MockSMTPClient struct {
	Addr string
}

func NewMockSMTPClient(addr string) *MockSMTPClient {
	return &MockSMTPClient{
		Addr: addr,
	}
}

func (c *MockSMTPClient) SendEmail(email MockEmail) error {
	var e error

	conn, e := net.Dial("tcp", c.Addr)
	if e != nil {
		return e
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	_, e = fmt.Fprintf(writer, "Title: %s\n", email.Title)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(writer, "Text: %s\n", email.Text)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(writer, "Html: %s\n", email.Html)
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
