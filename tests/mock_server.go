package api

import (
	"bufio"
	"net"
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
	s.Addr = listener.Addr().String()

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

func (s *MockSMTPServer) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	buf := bufio.NewReader(conn)
	var email MockEmail

	for {
		line, e := buf.ReadString('\n')
		if e != nil {
			return
		}
		// terminator
		if line == "." {
			break
		}
		if line[:6] == "Title:" {
			email.Title = line[7:]
		} else if line[:5] == "Text:" {
			email.Text = line[:6]
		} else if line[:5] == "Html:" {
			email.Html = line[:6]
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.Emails = append(s.Emails, email)
}

func (s *MockSMTPServer) ClearEmails() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.Emails = nil
}

func (s *MockSMTPServer) GetEmails() []MockEmail {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.Emails
}
