package service

import (
	"strings"
	"testing"
)

func TestRenderApprovalEmail(t *testing.T) {
	html := renderApprovalEmail("Zhang San", "gpu01.example.com", 20000, []int{20001, 20002}, "abc123", "Approved for research")

	checks := []string{"Container Application Approved", "Zhang San", "gpu01.example.com", "root", "20000", "abc123", "20001-20002", "Approved for research"}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("Expected email to contain %q", c)
		}
	}
}

func TestRenderApprovalEmailEscapesAdminNotes(t *testing.T) {
	html := renderApprovalEmail("Zhang San", "gpu01.example.com", 20000, []int{20001}, "abc123", "<script>alert(1)</script>")

	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Fatal("expected admin notes to be escaped")
	}
	if !strings.Contains(html, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Fatal("expected escaped admin notes to remain visible")
	}
}

func TestRenderRejectionEmail(t *testing.T) {
	html := renderRejectionEmail("Li Si", "GPU Server 01", "Ubuntu 22.04", "Insufficient resources")
	htmlLower := strings.ToLower(html)

	checks := []string{"Container Application Rejected", "Li Si", "GPU Server 01", "Ubuntu 22.04", "Insufficient resources", "could not be approved"}
	for _, c := range checks {
		if !strings.Contains(htmlLower, strings.ToLower(c)) {
			t.Errorf("Expected email to contain %q", c)
		}
	}
}

func TestRenderNewApplicationEmail(t *testing.T) {
	html := renderNewApplicationEmail("Wang Wu", "wang@example.com", "GPU Server 01", "Ubuntu CUDA")

	checks := []string{"New Container Application", "Wang Wu", "wang@example.com", "GPU Server 01", "Ubuntu CUDA"}
	for _, c := range checks {
		if !strings.Contains(html, c) {
			t.Errorf("Expected email to contain %q", c)
		}
	}
	if !strings.Contains(html, "mailto:wang@example.com") {
		t.Fatal("expected admin notification email to include mailto link")
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
