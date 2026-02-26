package handlers

import (
	"html/template"
	"net/http"

	"github.com/ftery0/ouath/server/store"
)

// loginPageData: login.html 템플릿에 넘겨줄 데이터 구조체
type loginPageData struct {
	ClientName  string
	State       string
	ClientID    string
	RedirectURI string
	Scope       string
	ErrorMsg    string
}

// AuthorizeHandler: GET /oauth/authorize 처리
func AuthorizeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		responseType := q.Get("response_type")
		clientID     := q.Get("client_id")
		redirectURI  := q.Get("redirect_uri")
		state        := q.Get("state")
		scope        := q.Get("scope")

		// redirect_uri 검증을 가장 먼저 해야 함 (오픈 리다이렉트 방지)
		client, ok := store.Clients.GetByClientID(clientID)
		if !ok || !containsURI(client.RedirectURIs, redirectURI) {
			renderError(w, tmpl, "유효하지 않은 client_id 또는 redirect_uri입니다")
			return
		}

		// response_type은 반드시 "code" 여야 함 (Authorization Code Flow)
		// 오류 시에는 redirect_uri로 에러 파라미터를 붙여서 돌려보냄
		if responseType != "code" {
			http.Redirect(w, r,
				redirectURI+"?error=unsupported_response_type&state="+state,
				http.StatusFound,
			)
			return
		}

		// 여기까지 왔으면 정상 — 로그인 페이지를 보여줌 (서비스명은 Client.Name 사용)
		data := loginPageData{
			ClientName:  client.Name,
			State:       state,
			ClientID:    clientID,
			RedirectURI: redirectURI,
			Scope:       scope,
		}
		if err := tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
			http.Error(w, "템플릿 렌더링 실패", http.StatusInternalServerError)
		}
	}
}

// renderError: redirect_uri 자체가 잘못된 경우 전용 에러 페이지 표시
// 이때는 리다이렉트 하면 안 되므로 에러 페이지를 직접 렌더링함
func renderError(w http.ResponseWriter, tmpl *template.Template, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	tmpl.ExecuteTemplate(w, "error.html", loginPageData{ErrorMsg: msg})
}

// containsURI: 허용된 URI 목록에 포함되어 있는지 검사
// 반드시 완전 일치(exact match)로 비교해야 함
// 접두사 일치(prefix match)를 쓰면 공격자가 example.com.evil.com 같은 걸 통과시킬 수 있음
func containsURI(uris []string, target string) bool {
	for _, u := range uris {
		if u == target {
			return true
		}
	}
	return false
}
