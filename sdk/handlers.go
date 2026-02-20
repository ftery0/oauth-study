package ouath

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

func (c *Client) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state, _ := randomHex(16)

	saveSession(w, session{
		State:      state,
		RedirectTo: r.URL.Query().Get("next"),
	})

	authorizeURL := c.cfg.ServerURL + "/oauth/authorize" +
		"?response_type=code" +
		"&client_id=" + c.cfg.ClientID +
		"&redirect_uri=" + c.cfg.RedirectURL +
		"&scope=" + strings.Join(c.cfg.Scopes, " ") +
		"&state=" + state

	http.Redirect(w, r, authorizeURL, http.StatusFound)
}

func (c *Client) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	sess, ok := loadSession(r)
	if !ok || sess.State != r.URL.Query().Get("state") {
		http.Error(w, "유효하지 않은 state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")

	accessToken, err := c.exchangeToken(code)
	if err != nil {
		http.Error(w, "토큰 교환 실패", http.StatusInternalServerError)
		return
	}

	userID, err := c.fetchUserID(accessToken)
	if err != nil {
		http.Error(w, "유저 정보 조회 실패", http.StatusInternalServerError)
		return
	}

	saveSession(w, session{
		UserID:      userID,
		AccessToken: accessToken,
	})

	redirectTo := sess.RedirectTo
	if redirectTo == "" {
		redirectTo = "/"
	}
	http.Redirect(w, r, redirectTo, http.StatusFound)
}

func (c *Client) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// 세션 쿠키를 만료시켜서 로그아웃 처리
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
	})
	http.Redirect(w, r, "/", http.StatusFound)
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return hex.EncodeToString(b), err
}
