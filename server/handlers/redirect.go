package handlers

import (
	"net/http"
	"net/url"
)

// safeOAuthRedirect: redirect_uri 에 OAuth 파라미터를 안전하게 결합한다.
//
// 기존 코드는 redirectURI + "?code=" + code + "&state=" + state 같이
// 문자열 결합을 썼는데, state 에 "abc&error=evil&iss=evil.com" 같은
// 값이 들어가면 querystring 이 corrupt (Open Redirect / parameter pollution).
//
// net/url.Values.Encode 를 거치면 모든 키/값이 안전하게 escape 된다.
func safeOAuthRedirect(w http.ResponseWriter, r *http.Request, redirectURI string, params map[string]string) {
	u, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusInternalServerError)
		return
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}
