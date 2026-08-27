package service

import (
	"crypto/tls"
	"fmt"
	"html/template"
	"log/slog"
	"mime"
	"net/smtp"
	"sort"
	"strconv"
	"strings"
)

type SMTPEmailService struct{ config *ConfigService }

const notificationSubject = "ServerDock · 容器通知"

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
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	message := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		username, to, encodedSubject, body,
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
<p style="margin:0 0 8px;color:#16803c;font-size:12px;font-weight:700;letter-spacing:.04em;">申请已批准</p>
<h1 style="margin:0;color:#1d1d1f;font-size:26px;line-height:1.25;letter-spacing:-.02em;">容器已创建</h1>
<p style="margin:12px 0 24px;color:#6e6e73;font-size:14px;line-height:1.7;">{{.ApplicantName}}，你的容器申请已通过审核。请使用以下信息连接容器。</p>
<div style="overflow:hidden;border:1px solid #d2d2d7;border-radius:12px;">
	  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="width:100%;table-layout:fixed;border-collapse:collapse;">
	    <tr><td width="76" style="width:76px;padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">服务器</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-family:SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.Server}}</td></tr>
	    <tr><td style="padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">用户</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-family:SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:13px;font-weight:600;overflow-wrap:anywhere;">root</td></tr>
	    <tr><td style="padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">密码</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-family:SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:13px;font-weight:600;word-break:break-all;">{{.Password}}</td></tr>
	    <tr><td style="padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">SSH 端口</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-family:SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.SSHPort}}</td></tr>
	    <tr><td style="padding:12px;color:#6e6e73;font-size:12px;white-space:nowrap;">额外端口</td><td style="padding:12px 14px;color:#1d1d1f;font-family:SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.ExtraPorts}}</td></tr>
  </table>
</div>
<p style="margin:12px 0 0;color:#8a5b00;font-size:12px;line-height:1.6;">连接密码仅在此邮件中提供，请妥善保管。</p>
<p style="margin:24px 0 8px;color:#6e6e73;font-size:11px;font-weight:700;letter-spacing:.04em;">SSH 命令</p>
<pre style="margin:0;padding:14px 16px;overflow-wrap:anywhere;border-radius:10px;background:#1d1d1f;color:#f5f5f7;font-family:SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:12px;line-height:1.6;white-space:pre-wrap;">ssh -p {{.SSHPort}} root@{{.Server}}</pre>` + emailEnd))

var rejectionEmail = template.Must(template.New("rejection").Parse(emailStart + `
<p style="margin:0 0 8px;color:#b42318;font-size:12px;font-weight:700;letter-spacing:.04em;">审核结果</p>
<h1 style="margin:0;color:#1d1d1f;font-size:26px;line-height:1.25;letter-spacing:-.02em;">容器申请未通过</h1>
<p style="margin:12px 0 24px;color:#6e6e73;font-size:14px;line-height:1.7;">{{.ApplicantName}}，本次容器申请未能通过审核。</p>
<div style="overflow:hidden;border:1px solid #d2d2d7;border-radius:12px;">
	  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="width:100%;table-layout:fixed;border-collapse:collapse;">
	    <tr><td width="76" style="width:76px;padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">服务器</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.Server}}</td></tr>
	    <tr><td style="padding:12px;color:#6e6e73;font-size:12px;white-space:nowrap;">镜像</td><td style="padding:12px 14px;color:#1d1d1f;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.Image}}</td></tr>
  </table>
</div>` + emailEnd))

var newApplicationEmail = template.Must(template.New("new").Parse(emailStart + `
<p style="margin:0 0 8px;color:#0066cc;font-size:12px;font-weight:700;letter-spacing:.04em;">待处理申请</p>
<h1 style="margin:0;color:#1d1d1f;font-size:26px;line-height:1.25;letter-spacing:-.02em;">收到新的容器申请</h1>
<p style="margin:12px 0 24px;color:#6e6e73;font-size:14px;line-height:1.7;">一份新的容器申请正在等待审核。</p>
<div style="overflow:hidden;border:1px solid #d2d2d7;border-radius:12px;">
	  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="width:100%;table-layout:fixed;border-collapse:collapse;">
	    <tr><td width="76" style="width:76px;padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">申请人</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.ApplicantName}}</td></tr>
	    <tr><td style="padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">邮箱</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;font-size:13px;font-weight:600;overflow-wrap:anywhere;"><a href="mailto:{{.Email}}" style="color:#0066cc;text-decoration:none;">{{.Email}}</a></td></tr>
	    <tr><td style="padding:12px;border-bottom:1px solid #e8e8ed;color:#6e6e73;font-size:12px;white-space:nowrap;">服务器</td><td style="padding:12px 14px;border-bottom:1px solid #e8e8ed;color:#1d1d1f;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.Server}}</td></tr>
	    <tr><td style="padding:12px;color:#6e6e73;font-size:12px;white-space:nowrap;">镜像</td><td style="padding:12px 14px;color:#1d1d1f;font-size:13px;font-weight:600;overflow-wrap:anywhere;">{{.Image}}</td></tr>
  </table>
</div>
{{if .Actions.ApproveURL}}
<p style="margin:24px 0 10px;color:#6e6e73;font-size:11px;font-weight:700;letter-spacing:.04em;">快速处理</p>
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="width:100%;border-collapse:collapse;">
  <tr>
    <td width="33.33%" style="padding-right:5px;"><a href="{{.Actions.IgnoreURL}}" style="display:block;padding:11px 8px;border:1px solid #d2d2d7;border-radius:9px;background:#ffffff;color:#3a3a3c;font-size:13px;font-weight:700;text-align:center;text-decoration:none;">忽略</a></td>
    <td width="33.33%" style="padding:0 5px;"><a href="{{.Actions.RejectURL}}" style="display:block;padding:11px 8px;border:1px solid #b42318;border-radius:9px;background:#ffffff;color:#b42318;font-size:13px;font-weight:700;text-align:center;text-decoration:none;">拒绝</a></td>
    <td width="33.33%" style="padding-left:5px;"><a href="{{.Actions.ApproveURL}}" style="display:block;padding:11px 8px;border:1px solid #16803c;border-radius:9px;background:#16803c;color:#ffffff;font-size:13px;font-weight:700;text-align:center;text-decoration:none;">批准</a></td>
  </tr>
</table>
<p style="margin:12px 0 0;color:#86868b;font-size:11px;line-height:1.6;">链接将在 7 天后失效，且申请处理后不可重复使用。链接可直接执行管理员操作，请勿转发此邮件。</p>
{{else}}
<p style="margin:20px 0 0;color:#6e6e73;font-size:12px;line-height:1.6;">请登录 ServerDock，在“申请审批”中处理此申请。配置公开访问地址和安全密钥后，后续通知会显示邮件审批按钮。</p>
{{end}}` + emailEnd))

const emailStart = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta name="color-scheme" content="light">
  <title>ServerDock 容器通知</title>
</head>
<body style="margin:0;padding:0;background-color:#f5f5f7;color:#1d1d1f;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Noto Sans SC',Arial,sans-serif;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" bgcolor="#f5f5f7" style="width:100%;border-collapse:collapse;background-color:#f5f5f7;">
    <tr>
      <td align="center" style="padding:32px 12px;">
        <table role="presentation" width="620" cellspacing="0" cellpadding="0" bgcolor="#ffffff" style="width:100%;max-width:620px;border:1px solid #d2d2d7;border-radius:16px;border-collapse:separate;background-color:#ffffff;overflow:hidden;">
          <tr>
            <td style="padding:24px 32px;border-bottom:1px solid #e8e8ed;">
              <table role="presentation" cellspacing="0" cellpadding="0" style="border-collapse:collapse;">
                <tr>
                  <td width="38" height="38" align="center" valign="middle" bgcolor="#0071e3" style="width:38px;height:38px;border-radius:10px;background-color:#0071e3;color:#ffffff;font-size:11px;font-weight:800;letter-spacing:-.02em;">SD</td>
                  <td style="padding-left:11px;">
                    <p style="margin:0;color:#1d1d1f;font-size:17px;font-weight:750;letter-spacing:-.02em;">ServerDock</p>
                    <p style="margin:2px 0 0;color:#86868b;font-size:10px;">基础设施控制台</p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding:30px 32px 32px;">`

const emailEnd = `
            </td>
          </tr>
          <tr>
            <td style="padding:17px 32px;border-top:1px solid #e8e8ed;color:#86868b;font-size:10px;line-height:1.6;">此邮件由 ServerDock 自动发送，请勿直接回复。</td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

func renderApprovalEmail(applicantName, server string, sshPort int, extraPorts []int, password string) string {
	return renderEmail(approvalEmail, struct {
		ApplicantName, Server, Password, ExtraPorts string
		SSHPort                                     int
	}{applicantName, server, password, formatPorts(extraPorts), sshPort})
}

func renderRejectionEmail(applicantName, server, image string) string {
	return renderEmail(rejectionEmail, struct{ ApplicantName, Server, Image string }{applicantName, server, image})
}

func renderNewApplicationEmail(applicantName, email, server, image string, actions emailActionLinks) string {
	return renderEmail(newApplicationEmail, struct {
		ApplicantName string
		Email         string
		Server        string
		Image         string
		Actions       emailActionLinks
	}{applicantName, email, server, image, actions})
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
