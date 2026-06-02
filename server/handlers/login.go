package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// LoginHandler: POST /oauth/login.
//
// 순서:
//  0. CSRF 검증 (Double-Submit Cookie)
//  1. 사용자 확인 + bcrypt — timing attack 방어 (미존재 ID 에도 dummy bcrypt)
//  2. 세션 고정 방어 — 기존 sid 명시적 폐기 후 새로 발급
//  3. IdP 세션 생성 + 쿠키 set (group_id 포함, cross-group silent 차단의 기준)
//  4. auth code 발급
//  5. redirect_uri 로 안전 redirect (Open Redirect 방어)
func LoginHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 0. CSRF 검증 — 어떤 처리보다 먼저
		if err := VerifyCSRFToken(r); err != nil {
			http.Error(w, "CSRF 검증 실패", http.StatusForbidden)
			return
		}

		id := r.FormValue("id")
		password := r.FormValue("password")
		state := r.FormValue("state")
		clientID := r.FormValue("client_id")
		redirectURI := r.FormValue("redirect_uri")
		scope := r.FormValue("scope")

		// 1. 사용자 확인 + bcrypt
		//    미존재 ID 도 dummy hash 로 동일 cost bcrypt → 응답 시간으로 enumeration 불가
		user, ok := models.TestUsers[id]
		hashToCompare := models.DummyPasswordHash
		if ok {
			hashToCompare = user.PasswordHash
		}
		err := bcrypt.CompareHashAndPassword(
			[]byte(hashToCompare),
			[]byte(password),
		)
		if !ok || err != nil {
			// 실패 — 폼 재렌더. 새 CSRF 토큰 발급 (재시도 위해)
			csrfToken, _ := NewCSRFToken(w)
			tmpl.ExecuteTemplate(w, "login.html", loginPageData{
				ClientName:  clientID,
				State:       state,
				ClientID:    clientID,
				RedirectURI: redirectURI,
				Scope:       scope,
				ErrorMsg:    "아이디 또는 비밀번호가 틀렸습니다",
				CSRFToken:   csrfToken,
			})
			return
		}

		// 2. 세션 고정 방어 — 기존 sid 가 있으면 명시적으로 폐기
		if oldSid, ok := GetIdPSessionID(r); ok {
			store.IdPSessions.Delete(oldSid)
		}

		// 3. 현재 로그인 그룹 ID 조회 (cross-group silent 차단 키)
		var groupID string
		if client, ok := store.Clients.GetByClientID(clientID); ok {
			groupID = client.GroupID
		}

		// 4. 새 IdP 세션 발급 + 쿠키 set
		sid, err := store.IdPSessions.Create(id, groupID)
		if err != nil {
			http.Error(w, "세션 생성 실패", http.StatusInternalServerError)
			return
		}
		if err := SetIdPSessionCookie(w, sid); err != nil {
			http.Error(w, "세션 쿠키 설정 실패", http.StatusInternalServerError)
			return
		}

		// 5. CSRF 쿠키 폐기 (토큰 재사용 방지)
		ClearCSRFToken(w)

		// 6. auth code 발급
		code, err := generateCode()
		if err != nil {
			http.Error(w, "서버 오류", http.StatusInternalServerError)
			return
		}
		store.AuthCodes.Store(code, models.AuthCode{
			Code:        code,
			ClientID:    clientID,
			UserID:      id,
			RedirectURI: redirectURI,
			Scope:       scope,
			ExpiresAt:   time.Now().Add(10 * time.Minute),
		})

		// 7. 안전 redirect (Open Redirect 방어)
		safeOAuthRedirect(w, r, redirectURI, map[string]string{
			"code":  code,
			"state": state,
		})
	}
}

func generateCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
