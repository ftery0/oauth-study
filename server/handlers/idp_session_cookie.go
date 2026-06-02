package handlers

import (
	"crypto/sha256"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"

	"github.com/ftery0/ouath/server/store"
)

// idpSessionCookieName: IdP 도메인 한정 세션 쿠키 이름.
const idpSessionCookieName = "idp_session"

// secureCookie: main 에서 IdPCookieInit 으로 한 번 주입.
// 핸들러 어디서든 이 인스턴스를 통해 sid 를 encode/decode 한다.
var secureCookie *securecookie.SecureCookie

// isProduction: main 에서 SetProduction 으로 설정. 쿠키 Secure 플래그 결정.
var isProduction bool

// SetProduction: main 에서 cfg.Env 기반으로 호출.
// 이후 SetIdPSessionCookie / NewCSRFToken 가 Secure 플래그를 자동으로 켠다.
func SetProduction(p bool) {
	isProduction = p
}

// IdPCookieInit: main 에서 secret 을 받아 SecureCookie 인스턴스 생성.
//
// secret 으로부터 두 개의 키를 도출한다:
//   - hashKey  (HMAC): sha256(secret)
//   - blockKey (AES) : sha256(secret + ".idp.block")
//
// 같은 secret 이라도 도메인 라벨이 다르므로 별도 키. 실 운영에서는
// HKDF 같은 KDF 권장이지만 학습 단계에서는 sha256 으로 충분.
func IdPCookieInit(secret string) {
	hash := sha256.Sum256([]byte(secret))
	block := sha256.Sum256([]byte(secret + ".idp.block"))
	secureCookie = securecookie.New(hash[:], block[:])
}

// SetIdPSessionCookie: sid 를 암호화/서명해서 쿠키로 굽는다.
//
// 속성:
//   - HttpOnly: JS 접근 차단 (XSS 시 토큰 직접 탈취 불가)
//   - SameSite=Lax: top-level redirect (다른 도메인에서 IdP 로의 navigation) 에서
//     쿠키가 전송돼야 silent SSO 가 동작. Strict 로 두면 깨진다.
//   - Path=/: IdP 전체 경로에서 사용
//   - Secure: production 일 때만 true (패키지 isProduction 참조)
func SetIdPSessionCookie(w http.ResponseWriter, sid string) error {
	encoded, err := secureCookie.Encode(idpSessionCookieName, sid)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     idpSessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isProduction,
		Expires:  time.Now().Add(store.IdPSessionTTL),
	})
	return nil
}

// GetIdPSessionID: 쿠키에서 sid 를 꺼낸다.
// 쿠키 없음 / 서명 실패 / decode 실패 → ("", false).
func GetIdPSessionID(r *http.Request) (string, bool) {
	c, err := r.Cookie(idpSessionCookieName)
	if err != nil {
		return "", false
	}
	var sid string
	if err := secureCookie.Decode(idpSessionCookieName, c.Value, &sid); err != nil {
		return "", false
	}
	return sid, true
}

// ClearIdPSessionCookie: 쿠키를 즉시 만료시킨다 (logout 시 호출).
func ClearIdPSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     idpSessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}
