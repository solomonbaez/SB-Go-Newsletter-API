package api

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type MockEmail struct {
	Title string
	Text  string
	Html  string
}

// in-memory SMTP server
type MockSMTPServer struct {
	Addr    string
	Emails  []MockEmail
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

func (s *MockSMTPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	buf := bufio.NewReader(conn)
	var email MockEmail
	var data string

	for {
		line, e := buf.ReadString('\n')
		if e != nil {
			return
		}
		// reamove \n
		line = strings.TrimRight(line, "\n")

		// terminator
		if line == "." {
			break
		}
		if strings.HasPrefix(line, "Title:") {
			email.Title = strings.TrimPrefix(line, "Title:")
		} else if strings.HasPrefix(line, "Text:") {
			email.Text = strings.TrimPrefix(line, "Text:")
		} else if strings.HasPrefix(line, "Html:") {
			email.Html = strings.TrimPrefix(line, "Html:")
		} else {
			fmt.Println("Unknown")
		}

		data += line + "\n"
	}

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

func (s *MockSMTPServer) GetEmails() []MockEmail {
	s.lock.Lock()
	defer s.lock.Unlock()
	fmt.Printf("Len: %v", len(s.Emails))
	return s.Emails
}
