package service

import (
	"crypto/tls"
	"fmt"
	"html"
	"log/slog"
	"net/smtp"
	"sort"
	"strings"
)

type SMTPEmailService struct {
	configService *ConfigService
}

func NewSMTPEmailService(configService *ConfigService) *SMTPEmailService {
	return &SMTPEmailService{configService: configService}
}

func (s *SMTPEmailService) SendAsync(to, subject, htmlBody string) {
	go func() {
		if err := s.Send(to, subject, htmlBody); err != nil {
			slog.Error("Failed to send email", "to", to, "error", err)
		}
	}()
}

func (s *SMTPEmailService) Send(to, subject, htmlBody string) error {
	if s.configService.Get("email_enabled") != "true" {
		slog.Info("Email sending disabled, skipping", "to", to, "subject", subject)
		return nil
	}

	host := s.configService.Get("smtp_host")
	port := s.configService.Get("smtp_port")
	username := s.configService.Get("smtp_username")
	password := s.configService.Get("smtp_password")
	useTLS := s.configService.Get("smtp_use_tls") == "true"

	addr := fmt.Sprintf("%s:%s", host, port)

	headers := map[string]string{
		"From":         username,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	auth := smtp.PlainAuth("", username, password, host)

	if useTLS {
		tlsConfig := &tls.Config{ServerName: host}
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("TLS dial failed: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, host)
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
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg.String()))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, username, []string{to}, []byte(msg.String()))
}

func escapeHTML(value string) string {
	return html.EscapeString(value)
}

func renderEmailShell(eyebrow, title, intro, content string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background:#f5f1e8;color:#2f2a24;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
  <div style="padding:20px 14px;background:
    radial-gradient(circle at top left, rgba(212,197,169,0.34), transparent 32%%),
    linear-gradient(180deg, #f7f4ed 0%%, #f2ede3 100%%);">
    <div style="max-width:640px;margin:0 auto;">
      <div style="margin-bottom:12px;color:#6e6253;font-size:11px;letter-spacing:0.14em;text-transform:uppercase;">%s</div>
      <div style="background:#fffdfa;border:1px solid #ddd4c7;border-radius:20px;padding:24px 24px 18px;box-shadow:0 12px 34px rgba(64,46,24,0.07);">
        <h1 style="margin:0 0 8px;font-size:24px;line-height:1.15;font-weight:600;color:#201b16;">%s</h1>
        <p style="margin:0 0 18px;font-size:14px;line-height:1.65;color:#564c40;">%s</p>
        %s
        <div style="margin-top:20px;padding-top:14px;border-top:1px solid #ece4d8;font-size:11px;line-height:1.65;color:#7a6e60;">
          Sent by ServerDock. This message contains environment details for your container workflow.
        </div>
      </div>
    </div>
  </div>
</body>
</html>`, escapeHTML(title), escapeHTML(eyebrow), escapeHTML(title), intro, content)
}

func renderInfoTable(rows [][2]string) string {
	var b strings.Builder
	b.WriteString(`<div style="border:1px solid #e8dfd2;border-radius:16px;overflow:hidden;background:#fcfaf6;">`)
	for i, row := range rows {
		border := ""
		if i > 0 {
			border = "border-top:1px solid #eee5d8;"
		}
		b.WriteString(fmt.Sprintf(
			`<div style="display:flex;gap:14px;align-items:flex-start;padding:12px 14px;%s"><div style="width:108px;flex-shrink:0;font-size:11px;letter-spacing:0.08em;text-transform:uppercase;color:#8c7f70;">%s</div><div style="font-size:13px;line-height:1.6;color:#2f2a24;">%s</div></div>`,
			border, escapeHTML(row[0]), row[1],
		))
	}
	b.WriteString(`</div>`)
	return b.String()
}

func renderCodeBlock(code string) string {
	return fmt.Sprintf(`<div style="margin-top:12px;padding:13px 14px;border-radius:16px;background:#23201c;color:#f7f2e8;font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,monospace;font-size:12px;line-height:1.65;overflow:auto;">%s</div>`, escapeHTML(code))
}

func formatPortRange(extraPorts []int) string {
	if len(extraPorts) == 0 {
		return "-"
	}

	ports := append([]int(nil), extraPorts...)
	sort.Ints(ports)
	start := ports[0]
	end := ports[0]
	for _, port := range ports[1:] {
		if port == end+1 {
			end = port
			continue
		}
		break
	}
	if start == end {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}

// RenderApprovalEmail renders HTML for an approved application.
func RenderApprovalEmail(applicantName, serverHostname string, sshPort int, extraPorts []int, password, adminNotes string) string {
	intro := fmt.Sprintf(
		`Hello %s,<br /><br />Your container request has been approved. Your environment is ready and the connection details are below.`,
		escapeHTML(applicantName),
	)

	rows := [][2]string{
		{"Server", escapeHTML(serverHostname)},
		{"User", "root"},
		{"Password", fmt.Sprintf(`<span style="font-family:'SFMono-Regular',Consolas,'Liberation Mono',Menlo,monospace;font-size:13px;padding:4px 8px;border-radius:10px;background:#f1eadf;">%s</span>`, escapeHTML(password))},
		{"SSH Port", fmt.Sprintf("%d", sshPort)},
		{"Extra Port", formatPortRange(extraPorts)},
	}

	var content strings.Builder
	content.WriteString(renderInfoTable(rows))
	content.WriteString(`<div style="margin-top:16px;">`)
	content.WriteString(`<div style="margin-bottom:8px;font-size:11px;letter-spacing:0.08em;text-transform:uppercase;color:#8c7f70;">Connect</div>`)
	content.WriteString(renderCodeBlock(fmt.Sprintf("ssh -p %d root@%s", sshPort, serverHostname)))
	content.WriteString(`</div>`)
	if strings.TrimSpace(adminNotes) != "" {
		content.WriteString(fmt.Sprintf(
			`<div style="margin-top:16px;padding:13px 14px;border-radius:16px;background:#f6f1e7;border:1px solid #e8dfd2;"><div style="margin-bottom:6px;font-size:11px;letter-spacing:0.08em;text-transform:uppercase;color:#8c7f70;">Admin Notes</div><div style="font-size:13px;line-height:1.65;color:#473d32;">%s</div></div>`,
			strings.ReplaceAll(escapeHTML(adminNotes), "\n", "<br />"),
		))
	}

	return renderEmailShell("Approval Update", "Container Application Approved", intro, content.String())
}

// RenderRejectionEmail renders HTML for a rejected application.
func RenderRejectionEmail(applicantName, serverHost, imageName, reason string) string {
	intro := fmt.Sprintf(
		`Hello %s,<br /><br />Your request for a container on <strong>%s</strong> using <strong>%s</strong> could not be approved at this time.`,
		escapeHTML(applicantName), escapeHTML(serverHost), escapeHTML(imageName),
	)

	content := fmt.Sprintf(
		`<div style="padding:14px 16px;border-radius:16px;background:#f8efe9;border:1px solid #edd8cc;"><div style="margin-bottom:6px;font-size:11px;letter-spacing:0.08em;text-transform:uppercase;color:#9a6e57;">Review Notes</div><div style="font-size:13px;line-height:1.65;color:#5a4336;">%s</div></div>`,
		strings.ReplaceAll(escapeHTML(strings.TrimSpace(reason)), "\n", "<br />"),
	)
	if strings.TrimSpace(reason) == "" {
		content = `<div style="padding:14px 16px;border-radius:16px;background:#f8efe9;border:1px solid #edd8cc;font-size:13px;line-height:1.65;color:#5a4336;">No additional notes were provided for this decision.</div>`
	}

	return renderEmailShell("Decision Update", "Container Application Rejected", intro, content)
}

// RenderNewApplicationEmail renders HTML for notifying admin of a new application.
func RenderNewApplicationEmail(applicantName, applicantEmail, serverHost, imageName string) string {
	intro := `A new container application has been submitted and is waiting for review in the admin panel.`
	content := renderInfoTable([][2]string{
		{"Applicant", escapeHTML(applicantName)},
		{"Email", fmt.Sprintf(`<a href="mailto:%s" style="color:#7a5638;text-decoration:none;">%s</a>`, escapeHTML(applicantEmail), escapeHTML(applicantEmail))},
		{"Server", escapeHTML(serverHost)},
		{"Image", escapeHTML(imageName)},
	})

	return renderEmailShell("Admin Notification", "New Container Application", intro, content)
}
