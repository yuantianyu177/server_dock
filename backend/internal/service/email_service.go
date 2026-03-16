package service

// EmailService defines the interface for sending emails.
type EmailService interface {
	SendAsync(to, subject, htmlBody string)
	Send(to, subject, htmlBody string) error
}
