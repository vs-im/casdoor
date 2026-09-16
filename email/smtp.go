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
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/gomail/v2"
	"golang.org/x/net/proxy"
)

// smtpSendTimeout bounds the whole Send(): dial, EHLO, STARTTLS, AUTH, MAIL/RCPT/DATA, QUIT.
const smtpSendTimeout = 30 * time.Second

// smtpDialTimeout bounds the TCP connect only.
const smtpDialTimeout = 10 * time.Second

type SmtpEmailProvider struct {
	Dialer *gomail.Dialer
	// Timeout for the whole Send call (0 means smtpSendTimeout).
	Timeout time.Duration
}

// GetSmtpLocalName returns the host name announced in EHLO/HELO.
// Order: config key "smtpLocalName" (env overrides app.conf) -> os.Hostname() -> "casdoor".
// An empty LocalName makes net/smtp send "EHLO localhost", which e.g. smtp-relay.gmail.com
// rejects with "421 4.7.0 Try again later, closing connection. (EHLO)".
func GetSmtpLocalName() string {
	return smtpLocalName(conf.GetConfigString("smtpLocalName"))
}

func smtpLocalName(configured string) string {
	name := strings.TrimSpace(configured)
	if name == "" {
		name, _ = os.Hostname()
		name = strings.TrimSpace(name)
	}
	if name == "" || strings.EqualFold(name, "localhost") {
		name = "casdoor"
	}
	return name
}

func NewSmtpEmailProvider(userName string, password string, host string, port int, typ string, sslMode string, enableProxy bool) *SmtpEmailProvider {
	dialer := gomail.NewDialer(host, port, userName, password)
	dialer.LocalName = GetSmtpLocalName()
	if typ == "SUBMAIL" {
		dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Handle SSL mode: "Auto" (or empty) means don't override gomail's default behavior
	// "Enable" means force SSL on, "Disable" means force SSL off
	if sslMode == "Enable" {
		dialer.SSL = true
	} else if sslMode == "Disable" {
		dialer.SSL = false
	}
	// If sslMode is "Auto" or empty, don't set dialer.SSL - let gomail decide based on port

	if enableProxy {
		socks5Proxy := conf.GetConfigString("socks5Proxy")
		if socks5Proxy != "" {
			dialer.SetSocks5Proxy(socks5Proxy)
		}
	}

	return &SmtpEmailProvider{Dialer: dialer, Timeout: smtpSendTimeout}
}

func (s *SmtpEmailProvider) Send(fromAddress string, fromName string, toAddresses []string, subject string, content string) error {
	message := gomail.NewMessage()

	message.SetAddressHeader("From", fromAddress, fromName)
	var addresses []string
	for _, address := range toAddresses {
		addresses = append(addresses, message.FormatAddress(address, ""))
	}
	message.SetHeader("To", addresses...)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", content)

	message.SkipUsernameCheck = true

	// NOTE: gomail's Dialer.DialAndSend is intentionally NOT used here. Its smtpSender.Send
	// re-dials and recurses without any limit whenever MAIL FROM returns io.EOF (which is
	// exactly what happens after the server answers 421 to EHLO and closes the connection),
	// producing an endless reconnect storm and a hanging HTTP request.
	session, err := s.dial()
	if err != nil {
		return err
	}
	defer session.Close()

	return gomail.Send(session, message)
}

// smtpSession is a gomail.SendCloser over a plain net/smtp client: one connection,
// one attempt, a single deadline for everything, no implicit reconnects.
type smtpSession struct {
	conn   net.Conn
	client *smtp.Client
}

func (s *SmtpEmailProvider) timeout() time.Duration {
	if s.Timeout > 0 {
		return s.Timeout
	}
	return smtpSendTimeout
}

func (s *SmtpEmailProvider) dial() (*smtpSession, error) {
	d := s.Dialer
	address := net.JoinHostPort(d.Host, fmt.Sprintf("%d", d.Port))

	var conn net.Conn
	var err error
	if d.Socks5Proxy != "" {
		var socksDialer proxy.Dialer
		socksDialer, err = proxy.SOCKS5("tcp", d.Socks5Proxy, nil, &net.Dialer{Timeout: smtpDialTimeout})
		if err != nil {
			return nil, err
		}
		conn, err = socksDialer.Dial("tcp", address)
	} else {
		conn, err = net.DialTimeout("tcp", address, smtpDialTimeout)
	}
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, errors.New("smtp: dial failed: connection is nil")
	}

	// One deadline for the entire SMTP conversation.
	_ = conn.SetDeadline(time.Now().Add(s.timeout()))

	if d.SSL {
		conn = tls.Client(conn, s.tlsConfig())
	}

	client, err := smtp.NewClient(conn, d.Host)
	if err != nil {
		conn.Close()
		return nil, err
	}
	session := &smtpSession{conn: conn, client: client}
	localName := smtpLocalName(d.LocalName)

	// Explicit EHLO with a real host name; the error (e.g. 421 on EHLO) is returned right away.
	// net/smtp falls back to HELO when EHLO fails; if the server already closed the socket
	// (Gmail relay: "421 4.7.0 Try again later, closing connection. (EHLO)") the reported error is io.EOF.
	if err = client.Hello(localName); err != nil {
		session.abort()
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("smtp: %s closed the connection on EHLO/HELO %s (e.g. 421 Try again later)", d.Host, localName)
		}
		return nil, fmt.Errorf("smtp: EHLO/HELO %s rejected by %s: %w", localName, d.Host, err)
	}

	if !d.SSL {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err = client.StartTLS(s.tlsConfig()); err != nil {
				session.abort()
				return nil, err
			}
		}
	}

	auth := d.Auth
	if auth == nil && d.Username != "" {
		if ok, auths := client.Extension("AUTH"); ok {
			switch {
			case strings.Contains(auths, "CRAM-MD5"):
				auth = smtp.CRAMMD5Auth(d.Username, d.Password)
			case strings.Contains(auths, "LOGIN") && !strings.Contains(auths, "PLAIN"):
				auth = &loginAuth{username: d.Username, password: d.Password, host: d.Host}
			default:
				auth = smtp.PlainAuth("", d.Username, d.Password, d.Host)
			}
		}
	}
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			session.abort()
			return nil, err
		}
	}

	return session, nil
}

func (s *SmtpEmailProvider) tlsConfig() *tls.Config {
	if s.Dialer.TLSConfig == nil {
		return &tls.Config{ServerName: s.Dialer.Host}
	}
	return s.Dialer.TLSConfig
}

// Send implements gomail.Sender: exactly one attempt, errors are returned to the caller.
func (s *smtpSession) Send(from string, to []string, msg io.WriterTo) error {
	if err := s.client.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err := s.client.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := s.client.Data()
	if err != nil {
		return err
	}
	if _, err = msg.WriteTo(w); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// Close sends QUIT; if the server is already gone the socket is closed anyway.
func (s *smtpSession) Close() error {
	err := s.client.Quit()
	if err != nil {
		s.abort()
	}
	return err
}

func (s *smtpSession) abort() {
	_ = s.client.Close()
	_ = s.conn.Close()
}

// loginAuth is an smtp.Auth that implements the LOGIN authentication mechanism
// (copied from gomail, where it is unexported).
type loginAuth struct {
	username string
	password string
	host     string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS {
		advertised := false
		for _, mechanism := range server.Auth {
			if mechanism == "LOGIN" {
				advertised = true
				break
			}
		}
		if !advertised {
			return "", nil, errors.New("smtp: unencrypted connection")
		}
	}
	if server.Name != a.host {
		return "", nil, errors.New("smtp: wrong host name")
	}
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch {
	case bytes.Equal(fromServer, []byte("Username:")):
		return []byte(a.username), nil
	case bytes.Equal(fromServer, []byte("Password:")):
		return []byte(a.password), nil
	default:
		return nil, fmt.Errorf("smtp: unexpected server challenge: %s", fromServer)
	}
}
