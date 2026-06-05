package handlers

import (
	"log/slog"
	"net/http"
	"os"
)

// auditLog: 토큰 발급 / 인증 성공·실패 / 폐기 등 보안 이벤트 단일 채널.
// stdout 으로 흐르므로 docker logs / journalctl 로 수집 가능.
var auditLog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

// AuditEvent: 정상 흐름 (200/302) — INFO.
func AuditEvent(r *http.Request, event string, attrs ...any) {
	auditLog.Info(event, append([]any{
		slog.String("ip", clientIP(r)),
		slog.String("ua", r.UserAgent()),
	}, attrs...)...)
}

// AuditWarn: 실패 / 거부 흐름 (4xx) — WARN.
func AuditWarn(r *http.Request, event string, attrs ...any) {
	auditLog.Warn(event, append([]any{
		slog.String("ip", clientIP(r)),
		slog.String("ua", r.UserAgent()),
	}, attrs...)...)
}
