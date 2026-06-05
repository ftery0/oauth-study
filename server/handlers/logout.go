package handlers

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/ftery0/ouath/server/models"
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

		// 3. post_logout_redirect_uri 안전 검증.
		// id_token_hint 가 있으면 그 client 의 MainURL/RedirectURIs 와 매칭.
		// 없으면 등록된 모든 client 를 훑어 매칭 (학습 단순화 — OIDC 표준은 정확 매칭 권장).
		if postLogoutURI != "" && isPostLogoutURIAllowed(postLogoutURI, redirectClient) {
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

// sameOrigin: 두 URL 이 같은 scheme+host+port 인가.
// 학습 편의: 등록된 MainURL 의 같은 origin 안 경로/쿼리는 post_logout_redirect_uri 로 허용.
func sameOrigin(a, b string) bool {
	au, err1 := url.Parse(a)
	bu, err2 := url.Parse(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return au.Scheme == bu.Scheme && au.Host == bu.Host
}

// isPostLogoutURIAllowed: 특정 client (또는 전체 client) 의 등록 URI 와 비교.
// hint 가 있으면 그 client 만, 없으면 모든 등록 client 의 MainURL / RedirectURIs / sameOrigin 매칭.
func isPostLogoutURIAllowed(uri, hintClientID string) bool {
	if hintClientID != "" {
		cl, ok := store.Clients.GetByClientID(hintClientID)
		return ok && clientAllowsLogoutURI(cl, uri)
	}
	for _, cl := range store.Clients.All() {
		if clientAllowsLogoutURI(cl, uri) {
			return true
		}
	}
	return false
}

func clientAllowsLogoutURI(cl *models.Client, uri string) bool {
	if cl.MainURL == uri || sameOrigin(cl.MainURL, uri) {
		return true
	}
	for _, u := range cl.RedirectURIs {
		if u == uri {
			return true
		}
	}
	return false
}
