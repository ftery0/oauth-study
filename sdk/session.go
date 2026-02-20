package ouath

import (
	"encoding/json"
	"net/http"
	"time"
)

const sessionCookie = "ouath_session"

type session struct {
	State       string
	RedirectTo  string
	UserID      string
	AccessToken string
}

func saveSession(w http.ResponseWriter, s session) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    string(b),
		Path:     "/",
		HttpOnly: true, // JS에서 쿠키 접근 차단
		Expires:  time.Now().Add(24 * time.Hour),
	})
	return nil
}

func loadSession(r *http.Request) (session, bool) {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return session{}, false
	}
	var s session
	if err := json.Unmarshal([]byte(c.Value), &s); err != nil {
		return session{}, false
	}
	return s, true
}
