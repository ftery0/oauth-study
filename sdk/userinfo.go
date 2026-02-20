package ouath

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) exchangeToken(code string) (string, error) {
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.cfg.RedirectURL},
	}

	req, err := http.NewRequest("POST", c.cfg.ServerURL+"/oauth/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// 서버가 Basic Auth로 클라이언트를 인증하므로 헤더에 담아 보냄
	req.SetBasicAuth(c.cfg.ClientID, c.cfg.ClientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("빈 access token")
	}
	return result.AccessToken, nil
}

func (c *Client) fetchUserID(accessToken string) (string, error) {
	req, _ := http.NewRequest("GET", c.cfg.ServerURL+"/oauth/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Sub string `json:"sub"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.Sub, nil
}

// UserFromSession: 핸들러 안에서 현재 로그인된 유저 ID를 꺼낼 때 사용
func UserFromSession(r *http.Request) (string, bool) {
	sess, ok := loadSession(r)
	if !ok || sess.UserID == "" {
		return "", false
	}
	return sess.UserID, true
}
