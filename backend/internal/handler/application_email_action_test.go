package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestApplicationHandlerEmailActionPageIsReadOnlyBootstrap(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewApplicationHandler(nil)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/applications/public/email-action?action=approve", nil)

	handler.EmailActionPage(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Header().Get("Cache-Control"), "no-store") {
		t.Fatalf("expected no-store response, got %q", recorder.Header().Get("Cache-Control"))
	}
	for _, check := range []string{"正在验证操作链接", "fetch(location.pathname", "method: 'POST'", "location.hash"} {
		if !strings.Contains(recorder.Body.String(), check) {
			t.Fatalf("email action page does not contain %q", check)
		}
	}
}
