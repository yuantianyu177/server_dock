package service

import (
	"strings"
	"testing"
)

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
