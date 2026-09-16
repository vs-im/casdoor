// Copyright 2023 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package email

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// fakeSmtp is a minimal scripted SMTP server. Its handler gets the connection's
// reader/writer and returns when the conversation is over; the connection is then closed.
type fakeSmtp struct {
	ln       net.Listener
	conns    atomic.Int32
	wg       sync.WaitGroup
	mu       sync.Mutex
	received []string
}

func startFakeSmtp(t *testing.T, handler func(s *fakeSmtp, r *bufio.Reader, w *bufio.Writer)) *fakeSmtp {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeSmtp{ln: ln}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			s.conns.Add(1)
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(20 * time.Second))
				handler(s, bufio.NewReader(c), bufio.NewWriter(c))
			}()
		}
	}()
	t.Cleanup(func() { ln.Close() })
	return s
}

func (s *fakeSmtp) hostPort(t *testing.T) (string, int) {
	t.Helper()
	host, portStr, _ := net.SplitHostPort(s.ln.Addr().String())
	port, _ := strconv.Atoi(portStr)
	return host, port
}

func reply(w *bufio.Writer, line string) {
	_, _ = w.WriteString(line + "\r\n")
	_ = w.Flush()
}

func readLine(r *bufio.Reader) (string, bool) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", false
	}
	return strings.TrimRight(line, "\r\n"), true
}

func newTestProvider(t *testing.T, s *fakeSmtp, timeout time.Duration) *SmtpEmailProvider {
	t.Helper()
	host, port := s.hostPort(t)
	p := NewSmtpEmailProvider("", "", host, port, "Default", "Disable", false)
	p.Dialer.LocalName = "casdoor-test"
	if timeout > 0 {
		p.Timeout = timeout
	}
	return p
}

// Reproduces smtp-relay.gmail.com: 220 banner, then 421 on EHLO and the server closes the socket.
func TestSmtpSend421OnEhloReturnsErrorOnce(t *testing.T) {
	s := startFakeSmtp(t, func(_ *fakeSmtp, r *bufio.Reader, w *bufio.Writer) {
		reply(w, "220 smtp-relay.fake ESMTP")
		if line, ok := readLine(r); ok && strings.HasPrefix(strings.ToUpper(line), "EHLO") {
			reply(w, "421-4.7.0 Try again later, closing connection. (EHLO)")
			reply(w, "421 4.7.0 fake")
		}
	})
	p := newTestProvider(t, s, 5*time.Second)

	start := time.Now()
	err := p.Send("noreply@example.com", "Casdoor", []string{"user@example.com"}, "subj", "<b>hi</b>")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error after 421 on EHLO")
	}
	if !strings.Contains(err.Error(), "EHLO") {
		t.Fatalf("error should mention EHLO, got: %v", err)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("Send took too long: %s", elapsed)
	}
	s.wg.Wait()
	if n := s.conns.Load(); n != 1 {
		t.Fatalf("expected exactly 1 connection (no reconnect loop), got %d", n)
	}
}

// Server accepts EHLO but drops the connection right after -> MAIL FROM gets io.EOF.
// gomail's smtpSender would re-dial recursively forever here; we must fail after one attempt.
func TestSmtpSendEofOnMailDoesNotReconnect(t *testing.T) {
	s := startFakeSmtp(t, func(_ *fakeSmtp, r *bufio.Reader, w *bufio.Writer) {
		reply(w, "220 fake ESMTP")
		if _, ok := readLine(r); ok {
			reply(w, "250 fake")
		}
		// close without reading MAIL FROM
	})
	p := newTestProvider(t, s, 5*time.Second)

	err := p.Send("noreply@example.com", "Casdoor", []string{"user@example.com"}, "subj", "body")
	if err == nil {
		t.Fatal("expected an error when the server closes before MAIL FROM")
	}
	s.wg.Wait()
	if n := s.conns.Load(); n != 1 {
		t.Fatalf("expected exactly 1 connection, got %d", n)
	}
}

// Server that answers the banner and then never says anything else -> Send must give up by Timeout.
func TestSmtpSendTimeout(t *testing.T) {
	s := startFakeSmtp(t, func(_ *fakeSmtp, r *bufio.Reader, w *bufio.Writer) {
		reply(w, "220 fake ESMTP")
		_, _ = readLine(r)
		time.Sleep(4 * time.Second)
	})
	p := newTestProvider(t, s, 1*time.Second)

	start := time.Now()
	err := p.Send("noreply@example.com", "Casdoor", []string{"user@example.com"}, "subj", "body")
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if elapsed > 3*time.Second {
		t.Fatalf("Send should have stopped after ~1s, took %s", elapsed)
	}
}

// Happy path: the EHLO carries the configured local name and the message is delivered.
func TestSmtpSendSuccess(t *testing.T) {
	s := startFakeSmtp(t, func(s *fakeSmtp, r *bufio.Reader, w *bufio.Writer) {
		reply(w, "220 fake ESMTP")
		inData := false
		for {
			line, ok := readLine(r)
			if !ok {
				return
			}
			s.mu.Lock()
			s.received = append(s.received, line)
			s.mu.Unlock()
			if inData {
				if line == "." {
					inData = false
					reply(w, "250 queued")
				}
				continue
			}
			switch strings.ToUpper(strings.Fields(line + " x")[0]) {
			case "EHLO":
				reply(w, "250-fake")
				reply(w, "250 8BITMIME")
			case "MAIL", "RCPT":
				reply(w, "250 OK")
			case "DATA":
				reply(w, "354 go")
				inData = true
			case "QUIT":
				reply(w, "221 bye")
				return
			default:
				reply(w, "500 what")
			}
		}
	})
	p := newTestProvider(t, s, 5*time.Second)

	if err := p.Send("noreply@example.com", "Casdoor", []string{"user@example.com"}, "subj", "body"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s.wg.Wait()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.received) == 0 || !strings.EqualFold(s.received[0], "EHLO casdoor-test") {
		t.Fatalf("first command should be 'EHLO casdoor-test', got %q", s.received)
	}
	joined := strings.Join(s.received, "\n")
	if !strings.Contains(joined, "MAIL FROM:<noreply@example.com>") || !strings.Contains(joined, "RCPT TO:<user@example.com>") {
		t.Fatalf("envelope missing in transcript:\n%s", joined)
	}
	if n := s.conns.Load(); n != 1 {
		t.Fatalf("expected exactly 1 connection, got %d", n)
	}
}

func TestSmtpLocalName(t *testing.T) {
	if got := smtpLocalName("mail.example.com"); got != "mail.example.com" {
		t.Fatalf("configured name should win, got %q", got)
	}
	if got := smtpLocalName("  "); got == "" || strings.EqualFold(got, "localhost") {
		t.Fatalf("empty config must fall back to hostname/casdoor, got %q", got)
	}
	if got := smtpLocalName("localhost"); got != "casdoor" {
		t.Fatalf("'localhost' must be replaced by 'casdoor', got %q", got)
	}
}
