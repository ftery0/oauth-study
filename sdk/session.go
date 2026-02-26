package ouath

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

const sessionCookie = "oauth_session"

func init() {
	gob.Register(session{})
}

type session struct {
	State       string
	RedirectTo  string
	UserID      string
	AccessToken string
}

// getSecureCookie: Client의 SessionSecret으로 SecureCookie 인스턴스 생성
// SessionSecret이 비어있으면 서명만 하는 fallback (하위호환)
func (c *Client) getSecureCookie() *securecookie.SecureCookie {
	if c.cfg.SessionSecret == "" {
		// 개발용: 고정 키 (프로덕션에서는 SessionSecret 필수)
		hashKey := []byte("oauth-dev-session-secret-32bytes!!")
		blockKey := []byte("oauth-dev-block-key-32bytes!!")
		return securecookie.New(hashKey, blockKey)
	}
	hashKey, blockKey := deriveKeys(c.cfg.SessionSecret)
	return securecookie.New(hashKey, blockKey)
}

func (c *Client) saveSession(w http.ResponseWriter, s session) error {
	encoded, err := c.getSecureCookie().Encode(sessionCookie, s)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // HTTPS 환경에서는 true로
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})
	return nil
}

func (c *Client) loadSession(r *http.Request) (session, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return session{}, false
	}
	var s session
	if err := c.getSecureCookie().Decode(sessionCookie, cookie.Value, &s); err != nil {
		return session{}, false
	}
	return s, true
}
