package handlers

import (
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

const (
	adminSessionCookieName = "admin_session"
	adminSessionTTL        = 12 * time.Hour
)

// adminSecureCookie: AdminInit 에서 주입.
var adminSecureCookie *securecookie.SecureCookie

// adminPasswordHash: bcrypt hash. main 에서 cfg.AdminPasswordHash 로 주입.
var adminPasswordHash string

// AdminInit: main 에서 한 번 호출. 비밀번호 hash + SecureCookie 키 설정.
func AdminInit(passwordHash, sessionSecret string) {
	adminPasswordHash = passwordHash
	hash := sha256.Sum256([]byte(sessionSecret))
	block := sha256.Sum256([]byte(sessionSecret + ".admin.block"))
	adminSecureCookie = securecookie.New(hash[:], block[:])
}

// setAdminSessionCookie: 어드민 게이트 통과 시 호출.
// 값은 단순 "ok" 마커 — 단일 비밀번호 모델이라 사용자 식별 정보 없음.
// 쿠키 Path 가 /admin 으로 한정되어 OAuth 흐름과 격리.
func setAdminSessionCookie(w http.ResponseWriter) error {
	encoded, err := adminSecureCookie.Encode(adminSessionCookieName, "ok")
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminSessionCookieName,
		Value:    encoded,
		Path:     "/admin",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isProduction,
		Expires:  time.Now().Add(adminSessionTTL),
	})
	return nil
}

// isAdminAuthed: 쿠키가 유효한지 확인.
func isAdminAuthed(r *http.Request) bool {
	c, err := r.Cookie(adminSessionCookieName)
	if err != nil {
		return false
	}
	var val string
	if err := adminSecureCookie.Decode(adminSessionCookieName, c.Value, &val); err != nil {
		return false
	}
	return val == "ok"
}

// clearAdminSessionCookie: 로그아웃 시 호출.
func clearAdminSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   adminSessionCookieName,
		Value:  "",
		Path:   "/admin",
		MaxAge: -1,
	})
}
