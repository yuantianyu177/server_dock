package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"serverdock/internal/dto"
	"serverdock/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *ApplicationHandler) EmailActionPage(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")
	c.Header("Content-Security-Policy", "default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	c.Header("Cross-Origin-Opener-Policy", "same-origin")
	c.Header("Permissions-Policy", "camera=(), geolocation=(), microphone=()")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(emailActionPageHTML))
}

func (h *ApplicationHandler) EmailAction(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("X-Content-Type-Options", "nosniff")

	var request dto.EmailApplicationActionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.EmailApplicationActionResponse{
			Status: "invalid_link", Title: "操作链接无效", Message: "链接中的安全凭证缺失，请从原始邮件重新点击。",
		})
		return
	}

	response, err := h.service.HandleEmailAction(request.Token)
	if err != nil {
		handleEmailActionError(c, err)
		return
	}

	result := dto.EmailApplicationActionResponse{Status: response.Status}
	switch response.Status {
	case "approved":
		result.Title = "申请已批准"
		result.Message = "容器已创建，系统已处理申请人的连接信息通知。"
	case "rejected":
		result.Title = "申请已拒绝"
		result.Message = "申请状态已更新，系统已处理申请人的审核结果通知。"
	case "ignored":
		result.Title = "申请已忽略"
		result.Message = "申请已静默归档，不会向申请人发送通知。"
	}
	c.JSON(http.StatusOK, result)
}

func handleEmailActionError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	result := dto.EmailApplicationActionResponse{
		Status: "failed", Title: "操作未完成", Message: "暂时无法处理该申请，请稍后从原始邮件重试，或进入管理后台查看。",
	}

	switch {
	case errors.Is(err, service.ErrInvalidEmailAction), errors.Is(err, service.ErrApplicationNotFound):
		status = http.StatusUnauthorized
		result = dto.EmailApplicationActionResponse{
			Status: "invalid_link", Title: "操作链接无效", Message: "此链接无效或已过期，请在管理后台处理该申请。",
		}
	case errors.Is(err, service.ErrApplicationNotPending):
		status = http.StatusConflict
		result = dto.EmailApplicationActionResponse{
			Status: "already_processed", Title: "申请已处理", Message: "这份申请已经完成审批，无需重复操作。",
		}
	case errors.Is(err, service.ErrContainerProvisioning):
		status = http.StatusBadGateway
		result = dto.EmailApplicationActionResponse{
			Status: "approval_failed", Title: "容器创建失败", Message: "申请仍保持待处理状态。请稍后从原始邮件重试，或进入管理后台检查服务器、镜像和端口配置。",
		}
	}

	if status >= http.StatusInternalServerError {
		slog.Error("Failed to handle application action from email", "error", err)
	}
	c.JSON(status, result)
}

const emailActionPageHTML = `<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <meta name="color-scheme" content="light">
  <title>邮件审批 · ServerDock</title>
  <style>
    :root{color-scheme:light;--paper:#fff;--canvas:#f5f5f7;--ink:#1d1d1f;--muted:#6e6e73;--line:#d2d2d7;--blue:#0071e3;--green:#16803c;--red:#b42318;--amber:#9a6700}
    *{box-sizing:border-box}
    body{margin:0;min-height:100vh;display:grid;place-items:center;padding:24px;background:var(--canvas);color:var(--ink);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","Noto Sans SC",Arial,sans-serif}
    .card{width:min(100%,520px);overflow:hidden;border:1px solid var(--line);border-radius:20px;background:var(--paper);box-shadow:0 24px 70px rgba(0,0,0,.08)}
    .brand{display:flex;align-items:center;gap:11px;padding:21px 26px;border-bottom:1px solid #e8e8ed}
    .brand-mark{display:grid;width:38px;height:38px;place-items:center;border-radius:10px;background:var(--blue);color:#fff;font-size:11px;font-weight:800;letter-spacing:-.02em}
    .brand strong{display:block;font-size:17px;letter-spacing:-.02em}
    .brand small{display:block;margin-top:2px;color:#86868b;font:10px/1.3 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;letter-spacing:.04em;text-transform:uppercase}
    .content{padding:30px 30px 32px}
    .action-rail{display:grid;grid-template-columns:repeat(3,1fr);gap:6px;margin-bottom:30px;padding:5px;border-radius:12px;background:#f2f2f4}
    .action-step{padding:7px 5px;border:1px solid transparent;border-radius:8px;color:#86868b;font-size:11px;font-weight:700;text-align:center}
    .action-step.active{border-color:#c7dcf4;background:#fff;color:var(--blue);box-shadow:0 2px 8px rgba(0,0,0,.05)}
    .action-step[data-action="ignore"].active{border-color:#c7c7cc;color:#3a3a3c}
    .action-step[data-action="reject"].active{border-color:#efc4bf;color:var(--red)}
    .action-step[data-action="approve"].active{border-color:#b7ddc4;color:var(--green)}
    .status-mark{display:grid;width:50px;height:50px;margin-bottom:18px;place-items:center;border-radius:15px;background:#eaf3fc;color:var(--blue)}
    .spinner{width:21px;height:21px;border:2px solid currentColor;border-right-color:transparent;border-radius:50%;animation:spin .8s linear infinite}
    .status-mark.done{background:#e9f6ed;color:var(--green)}
    .status-mark.warn{background:#fff5df;color:var(--amber)}
    .status-mark.error{background:#fff0ee;color:var(--red)}
    .status-symbol{font-size:23px;font-weight:800;line-height:1}
    .eyebrow{margin:0 0 7px;color:var(--blue);font:700 11px/1.3 ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;letter-spacing:.08em;text-transform:uppercase}
    h1{margin:0;font-size:27px;line-height:1.2;letter-spacing:-.035em}
    .message{margin:12px 0 0;color:var(--muted);font-size:14px;line-height:1.75}
    .foot{margin:24px 0 0;padding-top:18px;border-top:1px solid #e8e8ed;color:#86868b;font-size:11px;line-height:1.65}
    @keyframes spin{to{transform:rotate(360deg)}}
    @media(max-width:520px){body{padding:12px}.card{border-radius:16px}.brand{padding:18px 20px}.content{padding:25px 20px 27px}h1{font-size:24px}}
    @media(prefers-reduced-motion:reduce){.spinner{animation-duration:1.8s}}
  </style>
</head>
<body>
  <main class="card" aria-live="polite">
    <header class="brand"><span class="brand-mark" aria-hidden="true">SD</span><span><strong>ServerDock</strong><small>mail approval link</small></span></header>
    <section class="content">
      <div class="action-rail" aria-label="邮件审批操作">
        <span class="action-step" data-action="ignore">忽略</span>
        <span class="action-step" data-action="reject">拒绝</span>
        <span class="action-step" data-action="approve">批准</span>
      </div>
      <div id="status-mark" class="status-mark" aria-hidden="true"><span class="spinner"></span></div>
      <p id="eyebrow" class="eyebrow">邮件审批</p>
      <h1 id="title">正在验证操作链接</h1>
      <p id="message" class="message">验证通过后会立即处理，无需登录管理后台。</p>
      <p class="foot">完成后可直接关闭此页面。为保护审批权限，操作链接不会保存在此页面中。</p>
    </section>
  </main>
  <script>
    (() => {
      const action = new URLSearchParams(location.search).get('action')
      const token = new URLSearchParams(location.hash.slice(1)).get('token')
      const labels = { ignore: '忽略申请', reject: '拒绝申请', approve: '批准申请' }
      const statusActions = { ignored: 'ignore', rejected: 'reject', approved: 'approve' }
      const mark = document.getElementById('status-mark')
      const eyebrow = document.getElementById('eyebrow')
      const title = document.getElementById('title')
      const message = document.getElementById('message')
      document.querySelector('[data-action="' + (labels[action] ? action : '') + '"]')?.classList.add('active')
      history.replaceState(null, '', location.pathname + (labels[action] ? '?action=' + action : ''))

      const finish = (payload, ok) => {
        const warning = payload.status === 'already_processed'
        const completedAction = statusActions[payload.status]
        if (completedAction) {
          document.querySelectorAll('.action-step').forEach(step => step.classList.remove('active'))
          document.querySelector('[data-action="' + completedAction + '"]')?.classList.add('active')
        }
        mark.className = 'status-mark ' + (ok ? 'done' : warning ? 'warn' : 'error')
        mark.innerHTML = '<span class="status-symbol">' + (ok ? '✓' : warning ? '!' : '×') + '</span>'
        eyebrow.textContent = ok && completedAction ? labels[completedAction] : '邮件审批'
        title.textContent = payload.title || (ok ? '操作已完成' : '操作未完成')
        message.textContent = payload.message || '请稍后重试。'
        document.title = title.textContent + ' · ServerDock'
      }

      if (!token) {
        finish({ status: 'invalid_link', title: '操作链接无效', message: '链接中的安全凭证缺失，请从原始邮件重新点击；若仍然失败，请在管理后台处理。' }, false)
        return
      }

      fetch(location.pathname, {
        method: 'POST',
        credentials: 'omit',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token })
      }).then(async response => {
        const payload = await response.json().catch(() => ({}))
        finish(payload, response.ok)
      }).catch(() => {
        finish({ status: 'failed', title: '无法连接 ServerDock', message: '请检查网络后，从原始邮件重新点击此操作。' }, false)
      })
    })()
  </script>
</body>
</html>`
