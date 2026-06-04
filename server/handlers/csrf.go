package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"
)

const (
	csrfCookieName = "oauth_csrf"
	csrfTTL        = 10 * time.Minute
)

// ErrCSRFMismatch: CSRF token 검증 실패.
var ErrCSRFMismatch = errors.New("csrf token mismatch")

// NewCSRFToken: Double-Submit Cookie 방식.
// 같은 랜덤값을 (1) 응답 쿠키로 굽고 (2) 호출자가 폼 hidden 으로 박는다.
// 공격자는 다른 origin 에서 victim 의 cookie 를 읽을 수 없으므로
// 자기 form 에 같은 값을 넣지 못한다 → CSRF 차단.
//
// /oauth/authorize 가 로그인 폼을 렌더링하기 직전에 호출.
func NewCSRFToken(w http.ResponseWriter) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	http.SetCookie(w, &http.Cookie{
		Name:     csrfCookieName,
		Value:    token,
		Path:     "/oauth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isProduction,
		Expires:  time.Now().Add(csrfTTL),
	})
	return token, nil
}

// VerifyCSRFToken: /oauth/login POST 첫 단계에서 호출.
// form value 와 cookie value 가 같지 않으면 거부.
func VerifyCSRFToken(r *http.Request) error {
	formToken := r.FormValue("csrf_token")
	c, cookieErr := r.Cookie(csrfCookieName)
	cookieVal := ""
	if cookieErr == nil {
		cookieVal = c.Value
	}
	if formToken == "" || cookieErr != nil || cookieVal == "" || cookieVal != formToken {
		log.Printf("[csrf] mismatch path=%s formLen=%d cookieErr=%v cookieLen=%d match=%v",
			r.URL.Path, len(formToken), cookieErr, len(cookieVal), cookieVal == formToken)
		return ErrCSRFMismatch
	}
	return nil
}

// ClearCSRFToken: 로그인 성공 시 쿠키 폐기 (토큰 재사용 방지).
func ClearCSRFToken(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   csrfCookieName,
		Value:  "",
		Path:   "/oauth",
		MaxAge: -1,
	})
}
