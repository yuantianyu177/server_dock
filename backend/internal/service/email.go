package service

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"log/slog"
	"net/smtp"
	"sort"
	"strconv"
	"strings"
)

type SMTPEmailService struct{ config *ConfigService }

func NewSMTPEmailService(config *ConfigService) *SMTPEmailService {
	return &SMTPEmailService{config: config}
}

func (s *SMTPEmailService) SendAsync(to, subject, body string) {
	go func() {
		if err := s.Send(to, subject, body); err != nil {
			slog.Error("Failed to send email", "to", to, "error", err)
		}
	}()
}

func (s *SMTPEmailService) Send(to, subject, body string) error {
	if s.config.Get("email_enabled") != "true" {
		return nil
	}
	host, port := s.config.Get("smtp_host"), s.config.Get("smtp_port")
	username, password := s.config.Get("smtp_username"), s.config.Get("smtp_password")
	message := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		username, to, subject, body,
	))
	auth, address := smtp.PlainAuth("", username, password, host), host+":"+port
	if s.config.Get("smtp_use_tls") != "true" {
		return smtp.SendMail(address, auth, username, []string{to}, message)
	}

	connection, err := tls.Dial("tcp", address, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer connection.Close()
	client, err := smtp.NewClient(connection, host)
	if err != nil {
		return fmt.Errorf("SMTP client failed: %w", err)
	}
	defer client.Close()
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(username); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(message); err != nil {
		return err
	}
	return writer.Close()
}

var approvalEmail = template.Must(template.New("approval").Parse(emailStart + `
<h1>Container Application Approved</h1>
<p>Hello {{.ApplicantName}}, your container request has been approved.</p>
<dl><dt>Server</dt><dd>{{.Server}}</dd><dt>User</dt><dd>root</dd><dt>Password</dt><dd>{{.Password}}</dd><dt>SSH Port</dt><dd>{{.SSHPort}}</dd><dt>Extra Port</dt><dd>{{.ExtraPorts}}</dd></dl>
<pre>ssh -p {{.SSHPort}} root@{{.Server}}</pre>
{{if .Notes}}<h2>Admin Notes</h2><p>{{.Notes}}</p>{{end}}` + emailEnd))

var rejectionEmail = template.Must(template.New("rejection").Parse(emailStart + `
<h1>Container Application Rejected</h1>
<p>Hello {{.ApplicantName}}, your request for a container on <strong>{{.Server}}</strong> using <strong>{{.Image}}</strong> could not be approved.</p>
{{if .Notes}}<h2>Review Notes</h2><p>{{.Notes}}</p>{{else}}<p>No additional notes were provided.</p>{{end}}` + emailEnd))

var newApplicationEmail = template.Must(template.New("new").Parse(emailStart + `
<h1>New Container Application</h1>
<p>A new container application is waiting for review.</p>
<dl><dt>Applicant</dt><dd>{{.ApplicantName}}</dd><dt>Email</dt><dd><a href="mailto:{{.Email}}">{{.Email}}</a></dd><dt>Server</dt><dd>{{.Server}}</dd><dt>Image</dt><dd>{{.Image}}</dd></dl>` + emailEnd))

const emailStart = `<!doctype html><html lang="en"><meta charset="utf-8"><body style="font:14px sans-serif;line-height:1.6;color:#2f2a24;max-width:640px;margin:24px auto;padding:0 16px">`
const emailEnd = `<hr><small>Sent by ServerDock.</small></body></html>`

func renderApprovalEmail(applicantName, server string, sshPort int, extraPorts []int, password, notes string) string {
	return renderEmail(approvalEmail, struct {
		ApplicantName, Server, Password, ExtraPorts, Notes string
		SSHPort                                            int
	}{applicantName, server, password, formatPorts(extraPorts), notes, sshPort})
}

func renderRejectionEmail(applicantName, server, image, notes string) string {
	return renderEmail(rejectionEmail, struct{ ApplicantName, Server, Image, Notes string }{applicantName, server, image, notes})
}

func renderNewApplicationEmail(applicantName, email, server, image string) string {
	return renderEmail(newApplicationEmail, struct{ ApplicantName, Email, Server, Image string }{applicantName, email, server, image})
}

func renderEmail(t *template.Template, data any) string {
	var output strings.Builder
	if err := t.Execute(&output, data); err != nil {
		return ""
	}
	return output.String()
}

func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}
	ports = append([]int(nil), ports...)
	sort.Ints(ports)
	if ports[len(ports)-1]-ports[0] == len(ports)-1 {
		if len(ports) == 1 {
			return strconv.Itoa(ports[0])
		}
		return fmt.Sprintf("%d-%d", ports[0], ports[len(ports)-1])
	}
	values := make([]string, len(ports))
	for i, port := range ports {
		values[i] = strconv.Itoa(port)
	}
	return strings.Join(values, ", ")
}
