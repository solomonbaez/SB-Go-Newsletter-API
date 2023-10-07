package helpers

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/solomonbaez/SB-Go-Newsletter-API/api/models"
)

// in-memory SMTP server
type MockSMTPServer struct {
	Addr    string
	Emails  []models.Newsletter
	wg      sync.WaitGroup
	running bool
	lock    sync.Mutex
}

// builder
func NewMockSMTPServer() *MockSMTPServer {
	return &MockSMTPServer{}
}

func (s *MockSMTPServer) Start() {
	listener, e := net.Listen("tcp", ":0")
	if e != nil {
		panic(e)
	}

	s.lock.Lock()
	s.Addr = listener.Addr().String()
	s.running = true
	s.lock.Unlock()

	go func() {
		for {
			conn, e := listener.Accept()
			if e != nil {
				if !s.running {
					return
				}
				panic(e)
			}

			s.wg.Add(1)
			go s.handleConnection(conn)
		}
	}()
}

func (s *MockSMTPServer) Stop() {
	s.running = false
	s.wg.Wait()
}

func (s *MockSMTPServer) GetAddr() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Addr
}

// TODO bubble up errors
func (s *MockSMTPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	buf := bufio.NewReader(conn)
	var email models.Newsletter
	var body models.Body
	var data string

	for {
		line, e := buf.ReadString('\n')
		if e != nil {
			return
		}
		// remove \n
		line = strings.TrimRight(line, "\n")

		// terminator
		if line == "." {
			break
		}
		if strings.HasPrefix(line, "Recipient:") {
			recipient := strings.TrimPrefix(line, "Recipient: ")
			email.Recipient, e = models.ParseEmail(recipient)
			if e != nil {
				return
			}
		}
		if strings.HasPrefix(line, "Subject:") {
			body.Title = strings.TrimPrefix(line, "Subject: ")
		} else if strings.HasPrefix(line, "text/plain:") {
			body.Text = strings.TrimPrefix(line, "text/plain: ")
		} else if strings.HasPrefix(line, "text/html:") {
			body.Html = strings.TrimPrefix(line, "text/html: ")
		} else {
			return
		}

		data += line + "\n"
	}
	email.Content = &body

	s.lock.Lock()
	s.Emails = append(s.Emails, email)
	s.lock.Unlock()
	fmt.Printf("Recieved email:\n%s\n", data)
}

func (s *MockSMTPServer) ClearEmails() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Emails = nil
}

func (s *MockSMTPServer) GetEmails() []models.Newsletter {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Emails
}
