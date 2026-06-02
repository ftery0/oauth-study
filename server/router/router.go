package router

import (
	"html/template"
	"net/http"

	"github.com/ftery0/ouath/server/handlers"
)

// New: 완성된 핸들러는 실제 함수를, 아직 미구현은 stubHandler로 등록
// tmpl을 인자로 받는 이유: 핸들러들이 HTML을 렌더링하려면 템플릿이 필요하기 때문
func New(tmpl *template.Template) *http.ServeMux {
	mux := http.NewServeMux()

	// Go 1.22부터 "METHOD /path" 형식으로 메서드별 라우팅 가능
	mux.HandleFunc("GET /oauth/authorize", handlers.AuthorizeHandler(tmpl))
	mux.HandleFunc("POST /oauth/login", handlers.LoginHandler(tmpl))
	mux.HandleFunc("POST /oauth/token", handlers.TokenHandler)
	mux.HandleFunc("GET /oauth/userinfo", handlers.UserInfoHandler)
	// JWKS: 공개키 배포 엔드포인트 (클라이언트가 JWT 서명을 자체 검증할 때 사용)
	mux.HandleFunc("GET /oauth/jwks", handlers.JWKSHandler)

	// 어드민 (Phase 2-A): 단일 비밀번호 게이트 + read-only 시드 표시
	mux.HandleFunc("GET /admin/login", handlers.AdminLoginGetHandler(tmpl))
	mux.HandleFunc("POST /admin/login", handlers.AdminLoginPostHandler(tmpl))
	mux.HandleFunc("POST /admin/logout", handlers.AdminLogoutHandler)
	mux.Handle("GET /admin", handlers.RequireAdminAuth(handlers.AdminMainHandler(tmpl)))
	mux.Handle("GET /admin/clients/new", handlers.RequireAdminAuth(handlers.AdminClientNewFormHandler(tmpl)))
	mux.Handle("POST /admin/clients", handlers.RequireAdminAuth(handlers.AdminClientCreateHandler(tmpl)))
	mux.Handle("POST /admin/groups", handlers.RequireAdminAuth(http.HandlerFunc(handlers.AdminGroupCreateHandler)))
	mux.Handle("POST /admin/groups/{id}/policy", handlers.RequireAdminAuth(http.HandlerFunc(handlers.AdminGroupPolicyHandler)))

	return mux
}
