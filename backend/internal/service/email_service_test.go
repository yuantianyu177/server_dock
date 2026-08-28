package service

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSMTPEmailServiceRetriesTransientFailureWithSameMessage(t *testing.T) {
	service := newSMTPRetryTestService(t)
	var attempts int
	var messages [][]byte
	var delays []time.Duration
	service.deliver = func(_ smtpSettings, message []byte) error {
		attempts++
		messages = append(messages, append([]byte(nil), message...))
		if attempts < 3 {
			return &smtpDeliveryError{stage: "connection", retryable: true, err: io.EOF}
		}
		return nil
	}
	service.sleep = func(delay time.Duration) { delays = append(delays, delay) }

	if err := service.Send("yty@example.com", notificationSubject, "<p>yty</p>"); err != nil {
		t.Fatalf("send after transient failures: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if len(delays) != 2 || delays[0] != time.Second || delays[1] != 2*time.Second {
		t.Fatalf("unexpected retry delays: %v", delays)
	}
	for _, message := range messages[1:] {
		if !bytes.Equal(messages[0], message) {
			t.Fatal("all retry attempts must reuse the same message and Message-ID")
		}
	}
	if !bytes.Contains(messages[0], []byte("Message-ID: <serverdock-")) {
		t.Fatal("email is missing a stable Message-ID")
	}
}

func TestSMTPEmailServiceStopsAfterMaximumAttempts(t *testing.T) {
	service := newSMTPRetryTestService(t)
	attempts := 0
	service.deliver = func(_ smtpSettings, _ []byte) error {
		attempts++
		return &smtpDeliveryError{stage: "connection", retryable: true, err: io.EOF}
	}
	service.sleep = func(time.Duration) {}

	err := service.Send("yty@example.com", notificationSubject, "<p>yty</p>")
	if err == nil || !strings.Contains(err.Error(), "after 10 attempt(s)") {
		t.Fatalf("expected bounded retry error, got %v", err)
	}
	if attempts != smtpMaxAttempts {
		t.Fatalf("retry loop made %d attempts, want %d", attempts, smtpMaxAttempts)
	}
}

func TestSMTPRetryDelayIsCapped(t *testing.T) {
	tests := []struct {
		failedAttempt int
		want          time.Duration
	}{
		{failedAttempt: 0, want: 0},
		{failedAttempt: 1, want: time.Second},
		{failedAttempt: 2, want: 2 * time.Second},
		{failedAttempt: 5, want: 16 * time.Second},
		{failedAttempt: 6, want: smtpMaxRetryDelay},
		{failedAttempt: 9, want: smtpMaxRetryDelay},
	}
	for _, test := range tests {
		if got := smtpRetryDelay(test.failedAttempt); got != test.want {
			t.Errorf("smtpRetryDelay(%d) = %s, want %s", test.failedAttempt, got, test.want)
		}
	}
}

func TestSMTPEmailServiceDoesNotRetryAmbiguousCommitFailure(t *testing.T) {
	service := newSMTPRetryTestService(t)
	attempts := 0
	service.deliver = func(_ smtpSettings, _ []byte) error {
		attempts++
		return &smtpDeliveryError{stage: "message commit", retryable: false, err: io.EOF}
	}
	service.sleep = func(time.Duration) { t.Fatal("ambiguous commit failure must not be retried") }

	if err := service.Send("yty@example.com", notificationSubject, "<p>yty</p>"); err == nil {
		t.Fatal("expected ambiguous commit failure")
	}
	if attempts != 1 {
		t.Fatalf("ambiguous commit failure made %d attempts, want 1", attempts)
	}
}

func TestSMTPEmailServiceDoesNotRetryPermanentSMTPError(t *testing.T) {
	service := newSMTPRetryTestService(t)
	attempts := 0
	service.deliver = func(_ smtpSettings, _ []byte) error {
		attempts++
		return smtpCommandDeliveryError("auth", &textproto.Error{Code: 535, Msg: "authentication failed"})
	}
	service.sleep = func(time.Duration) { t.Fatal("permanent SMTP error must not be retried") }

	if err := service.Send("yty@example.com", notificationSubject, "<p>yty</p>"); err == nil {
		t.Fatal("expected permanent SMTP error")
	}
	if attempts != 1 {
		t.Fatalf("permanent SMTP error made %d attempts, want 1", attempts)
	}
}

func newSMTPRetryTestService(t *testing.T) *SMTPEmailService {
	t.Helper()
	config := setupConfigService(t)
	values := map[string]string{
		"email_enabled": "true",
		"smtp_host":     "smtp.example.com",
		"smtp_port":     "587",
		"smtp_username": "sender@example.com",
		"smtp_password": "secret",
		"smtp_use_tls":  "true",
	}
	for key, value := range values {
		if err := config.Set(key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}
	return NewSMTPEmailService(config)
}

func TestDialSMTPWithTLSUsesSTARTTLSOutsidePort465(t *testing.T) {
	serverCertificate, clientTLS := newSMTPTestCertificate(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	serverResult := serveStartTLSSMTP(listener, serverCertificate)

	client, err := dialSMTPWithTLS(listener.Addr().String(), "localhost", "587", clientTLS)
	if err != nil {
		t.Fatalf("dial SMTP with STARTTLS: %v", err)
	}
	if _, ok := client.TLSConnectionState(); !ok {
		t.Fatal("expected SMTP connection to be protected by STARTTLS")
	}
	client.Close()
	if err := <-serverResult; err != nil {
		t.Fatal(err)
	}
}

func TestDialSMTPWithTLSUsesImplicitTLSOnPort465(t *testing.T) {
	serverCertificate, clientTLS := newSMTPTestCertificate(t)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	serverResult := serveImplicitTLSSMTP(listener, serverCertificate)

	client, err := dialSMTPWithTLS(listener.Addr().String(), "localhost", "465", clientTLS)
	if err != nil {
		t.Fatalf("dial SMTP with implicit TLS: %v", err)
	}
	if err := client.Hello("localhost"); err != nil {
		t.Fatalf("SMTP hello: %v", err)
	}
	if _, ok := client.TLSConnectionState(); !ok {
		t.Fatal("expected SMTP connection to use implicit TLS")
	}
	client.Close()
	if err := <-serverResult; err != nil {
		t.Fatal(err)
	}
}

func TestSMTPEmailServiceDoesNotRetryAfterServerAcceptsMessage(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()
	serverResult := serveSMTPAcceptThenDropOnQuit(listener)

	service := newSMTPRetryTestService(t)
	port := listener.Addr().(*net.TCPAddr).Port
	for key, value := range map[string]string{
		"smtp_host":    "localhost",
		"smtp_port":    strconv.Itoa(port),
		"smtp_use_tls": "false",
	} {
		if err := service.config.Set(key, value); err != nil {
			t.Fatalf("set %s: %v", key, err)
		}
	}
	attempts := 0
	service.deliver = func(settings smtpSettings, message []byte) error {
		attempts++
		return deliverSMTPMessage(settings, message)
	}
	service.sleep = func(time.Duration) { t.Fatal("accepted message must not be retried") }

	if err := service.Send("yty@example.com", notificationSubject, "<p>yty</p>"); err != nil {
		t.Fatalf("server accepted the message: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("accepted message was sent %d times, want 1", attempts)
	}
	if err := <-serverResult; err != nil {
		t.Fatal(err)
	}
}

func newSMTPTestCertificate(t *testing.T) (tls.Certificate, *tls.Config) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
	}
	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	certificate := tls.Certificate{Certificate: [][]byte{certificateDER}, PrivateKey: privateKey}
	parsedCertificate, err := x509.ParseCertificate(certificateDER)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(parsedCertificate)
	return certificate, &tls.Config{ServerName: "localhost", RootCAs: roots, MinVersion: tls.VersionTLS12}
}

func serveStartTLSSMTP(listener net.Listener, certificate tls.Certificate) <-chan error {
	result := make(chan error, 1)
	go func() {
		result <- func() error {
			connection, err := listener.Accept()
			if err != nil {
				return err
			}
			defer connection.Close()
			plain := textproto.NewConn(connection)
			if err := plain.PrintfLine("220 localhost ESMTP"); err != nil {
				return err
			}
			if err := expectSMTPCommand(plain, "EHLO "); err != nil {
				return err
			}
			if err := plain.PrintfLine("250-localhost"); err != nil {
				return err
			}
			if err := plain.PrintfLine("250 STARTTLS"); err != nil {
				return err
			}
			if err := expectSMTPCommand(plain, "STARTTLS"); err != nil {
				return err
			}
			if err := plain.PrintfLine("220 ready for TLS"); err != nil {
				return err
			}

			secureConnection := tls.Server(connection, &tls.Config{
				Certificates: []tls.Certificate{certificate},
				MinVersion:   tls.VersionTLS12,
			})
			secure := textproto.NewConn(secureConnection)
			if err := expectSMTPCommand(secure, "EHLO "); err != nil {
				return err
			}
			return secure.PrintfLine("250 localhost")
		}()
	}()
	return result
}

func serveImplicitTLSSMTP(listener net.Listener, certificate tls.Certificate) <-chan error {
	result := make(chan error, 1)
	go func() {
		result <- func() error {
			connection, err := listener.Accept()
			if err != nil {
				return err
			}
			defer connection.Close()
			secureConnection := tls.Server(connection, &tls.Config{
				Certificates: []tls.Certificate{certificate},
				MinVersion:   tls.VersionTLS12,
			})
			secure := textproto.NewConn(secureConnection)
			if err := secure.PrintfLine("220 localhost ESMTP"); err != nil {
				return err
			}
			if err := expectSMTPCommand(secure, "EHLO "); err != nil {
				return err
			}
			return secure.PrintfLine("250 localhost")
		}()
	}()
	return result
}

func serveSMTPAcceptThenDropOnQuit(listener net.Listener) <-chan error {
	result := make(chan error, 1)
	go func() {
		result <- func() error {
			connection, err := listener.Accept()
			if err != nil {
				return err
			}
			defer connection.Close()
			server := textproto.NewConn(connection)
			if err := server.PrintfLine("220 localhost ESMTP"); err != nil {
				return err
			}
			if err := expectSMTPCommand(server, "EHLO "); err != nil {
				return err
			}
			if err := server.PrintfLine("250-localhost"); err != nil {
				return err
			}
			if err := server.PrintfLine("250 AUTH PLAIN"); err != nil {
				return err
			}
			if err := expectSMTPCommand(server, "AUTH PLAIN "); err != nil {
				return err
			}
			if err := server.PrintfLine("235 authenticated"); err != nil {
				return err
			}
			if err := expectSMTPCommand(server, "MAIL FROM:"); err != nil {
				return err
			}
			if err := server.PrintfLine("250 sender accepted"); err != nil {
				return err
			}
			if err := expectSMTPCommand(server, "RCPT TO:"); err != nil {
				return err
			}
			if err := server.PrintfLine("250 recipient accepted"); err != nil {
				return err
			}
			if err := expectSMTPCommand(server, "DATA"); err != nil {
				return err
			}
			if err := server.PrintfLine("354 send message"); err != nil {
				return err
			}
			message, err := io.ReadAll(server.DotReader())
			if err != nil {
				return err
			}
			if !bytes.Contains(message, []byte("<p>yty</p>")) {
				return fmt.Errorf("SMTP server received unexpected message: %q", message)
			}
			if err := server.PrintfLine("250 queued"); err != nil {
				return err
			}
			// Read QUIT, then close without its 221 response. The message was already
			// accepted and retrying at this point would create a duplicate.
			return expectSMTPCommand(server, "QUIT")
		}()
	}()
	return result
}

func expectSMTPCommand(connection *textproto.Conn, prefix string) error {
	line, err := connection.ReadLine()
	if err != nil {
		return err
	}
	if !strings.HasPrefix(line, prefix) {
		return fmt.Errorf("expected SMTP command %q, got %q", prefix, line)
	}
	return nil
}

func TestRenderApprovalEmail(t *testing.T) {
	html := renderApprovalEmail("Zhang San", "gpu01.example.com", 20000, []int{20001, 20002}, "abc123")

	checks := []string{"容器已创建", "申请已批准", "Zhang San", "gpu01.example.com", "root", "20000", "abc123", "20001-20002"}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("Expected email to contain %q", c)
		}
	}
}

func TestRenderRejectionEmail(t *testing.T) {
	html := renderRejectionEmail("Li Si", "GPU Server 01", "Ubuntu 22.04")

	checks := []string{"容器申请未通过", "审核结果", "Li Si", "GPU Server 01", "Ubuntu 22.04"}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("Expected email to contain %q", c)
		}
	}
}

func TestRenderNewApplicationEmail(t *testing.T) {
	html := renderNewApplicationEmail("Wang Wu", "wang@example.com", "GPU Server 01", "Ubuntu CUDA", emailActionLinks{})

	checks := []string{"收到新的容器申请", "待处理申请", "Wang Wu", "wang@example.com", "GPU Server 01", "Ubuntu CUDA"}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("Expected email to contain %q", c)
		}
	}
	if !strings.Contains(html, "mailto:wang@example.com") {
		t.Fatal("expected admin notification email to include mailto link")
	}
}

func TestRenderNewApplicationEmailIncludesActionButtons(t *testing.T) {
	html := renderNewApplicationEmail(
		"Wang Wu",
		"wang@example.com",
		"GPU Server 01",
		"Ubuntu CUDA",
		emailActionLinks{
			IgnoreURL:  "https://dock.example.com/action?type=ignore#token=ignore-token",
			RejectURL:  "https://dock.example.com/action?type=reject#token=reject-token",
			ApproveURL: "https://dock.example.com/action?type=approve#token=approve-token",
		},
	)

	checks := []string{
		">忽略</a>", ">拒绝</a>", ">批准</a>",
		"type=ignore#token=ignore-token",
		"type=reject#token=reject-token",
		"type=approve#token=approve-token",
		"链接将在 7 天后失效",
	}
	for _, check := range checks {
		if !strings.Contains(html, check) {
			t.Errorf("expected action email to contain %q", check)
		}
	}
}

func TestEmailTemplatesShareServerDockVisualSystem(t *testing.T) {
	emails := map[string]string{
		"approval":  renderApprovalEmail("Zhang San", "gpu01.example.com", 20000, nil, "abc123"),
		"rejection": renderRejectionEmail("Li Si", "GPU Server 01", "Ubuntu 22.04"),
		"new":       renderNewApplicationEmail("Wang Wu", "wang@example.com", "GPU Server 01", "Ubuntu CUDA", emailActionLinks{}),
	}
	checks := []string{
		`lang="zh-CN"`,
		`background-color:#f5f5f7`,
		`background-color:#0071e3`,
		`border-radius:16px`,
		`table-layout:fixed`,
		`width="76"`,
		`基础设施控制台`,
		`此邮件由 ServerDock 自动发送`,
	}
	for name, html := range emails {
		for _, check := range checks {
			if !strings.Contains(html, check) {
				t.Errorf("%s email does not contain shared design marker %q", name, check)
			}
		}
		if strings.Contains(html, `width:116px`) {
			t.Errorf("%s email still uses the oversized label column", name)
		}
	}
}

type MockEmailServiceImpl struct {
	AsyncCalls int
	SentEmails []struct {
		To      string
		Subject string
		Body    string
	}
}

func (m *MockEmailServiceImpl) SendAsync(to, subject, htmlBody string) {
	m.AsyncCalls++
	m.SentEmails = append(m.SentEmails, struct {
		To      string
		Subject string
		Body    string
	}{to, subject, htmlBody})
}
