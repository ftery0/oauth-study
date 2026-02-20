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

	return mux
}
