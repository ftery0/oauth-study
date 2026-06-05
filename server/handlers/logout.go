package handlers

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/ftery0/ouath/server/store"
	"github.com/ftery0/ouath/server/token"
)

// LogoutHandler: GET/POST /oauth/logout (OIDC end_session_endpoint).
//
// 흐름:
//  1. id_token_hint (있으면) 파싱해서 어떤 client 인지 파악 (선택적 — RP 식별)
//  2. IdP 세션 폐기 + 쿠키 만료
//  3. post_logout_redirect_uri 가 있고, id_token_hint 의 client 가 등록한 RedirectURIs / MainURL 과
//     일치하면 그 URL 로 redirect (state 동승). 매칭 실패면 안전한 안내 페이지 렌더.
func LogoutHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		idTokenHint := q.Get("id_token_hint")
		postLogoutURI := q.Get("post_logout_redirect_uri")
		state := q.Get("state")

		// 1. IdP 세션 폐기
		if sid, ok := GetIdPSessionID(r); ok {
			store.IdPSessions.Delete(sid)
		}
		ClearIdPSessionCookie(w)

		// 2. id_token_hint 가 있으면 aud 로 client 식별 (RP redirect URI 검증용)
		var redirectClient string
		if idTokenHint != "" {
			if c, err := token.ParseIDToken(idTokenHint); err == nil && len(c.Audience) > 0 {
				redirectClient = c.Audience[0]
			}
		}

		// 3. post_logout_redirect_uri 안전 검증
		if postLogoutURI != "" && redirectClient != "" {
			if cl, ok := store.Clients.GetByClientID(redirectClient); ok {
				allowed := containsURI(cl.RedirectURIs, postLogoutURI) || cl.MainURL == postLogoutURI
				if allowed {
					target := postLogoutURI
					if state != "" {
						sep := "?"
						if containsQuery(target) {
							sep = "&"
						}
						target += sep + "state=" + url.QueryEscape(state)
					}
					http.Redirect(w, r, target, http.StatusFound)
					return
				}
			}
		}

		// 검증 실패하거나 redirect 안 줬을 때: 안내 페이지
		tmpl.ExecuteTemplate(w, "error.html", loginPageData{
			ErrorMsg: "로그아웃되었습니다.",
		})
	}
}

func containsQuery(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '?' {
			return true
		}
	}
	return false
}
