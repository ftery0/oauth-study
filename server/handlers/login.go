package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// LoginHandler: POST /oauth/login.
//
// Phase-R R-3: TestUsers map 대신 store.Users (Postgres) 조회.
//
// 순서:
//  0. CSRF 검증 (Double-Submit Cookie)
//  1. 사용자 확인 + bcrypt — timing attack 방어 (미존재 username 에도 dummy bcrypt)
//  2. 세션 고정 방어 — 기존 sid 명시적 폐기 후 새로 발급
//  3. IdP 세션 생성 + 쿠키 set
//  4. auth code 발급
//  5. redirect_uri 로 안전 redirect (Open Redirect 방어)
func LoginHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 0. CSRF 검증 — 어떤 처리보다 먼저
		if err := VerifyCSRFToken(r); err != nil {
			http.Error(w, "CSRF 검증 실패", http.StatusForbidden)
			return
		}

		username := r.FormValue("id") // form field name 은 그대로 (login.html 유지)
		password := r.FormValue("password")
		state := r.FormValue("state")
		clientID := r.FormValue("client_id")
		redirectURI := r.FormValue("redirect_uri")
		scope := r.FormValue("scope")
		codeChallenge := r.FormValue("code_challenge")
		codeChallengeMethod := r.FormValue("code_challenge_method")
		nonce := r.FormValue("nonce")

		// 1. 사용자 확인 + bcrypt
		//    미존재 username 도 dummy hash 로 동일 cost bcrypt → 응답 시간으로 enumeration 불가
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		user, err := store.Users.GetByUsername(ctx, username)
		hashToCompare := models.DummyPasswordHash
		found := err == nil
		if found {
			hashToCompare = user.PasswordHash
		}

		bcryptErr := bcrypt.CompareHashAndPassword(
			[]byte(hashToCompare),
			[]byte(password),
		)
		if !found || bcryptErr != nil {
			AuditWarn(r, "login.failed", "username", username, "reason", map[bool]string{true: "bad_password", false: "user_not_found"}[found])
			// 미존재 user 케이스에서 ErrUserNotFound 외 다른 에러는 로그
			if err != nil && !errors.Is(err, store.ErrUserNotFound) {
				// DB 장애 등은 500 도 합리적이지만 학습 단계 사용자 enumeration 방어 우선
				_ = err
			}

			csrfToken, _ := NewCSRFToken(w)
			tmpl.ExecuteTemplate(w, "login.html", loginPageData{
				ClientName:          clientID,
				State:               state,
				ClientID:            clientID,
				RedirectURI:         redirectURI,
				Scope:               scope,
				ErrorMsg:            "아이디 또는 비밀번호가 틀렸습니다",
				CSRFToken:           csrfToken,
				CodeChallenge:       codeChallenge,
				CodeChallengeMethod: codeChallengeMethod,
				Nonce:               nonce,
			})
			return
		}

		// 2. 세션 고정 방어 — 기존 sid 가 있으면 명시적으로 폐기
		if oldSid, ok := GetIdPSessionID(r); ok {
			store.IdPSessions.Delete(oldSid)
		}

		// 3. 새 IdP 세션 발급 — UserID 는 DB 의 UUID
		sid, err := store.IdPSessions.Create(user.ID)
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
			Code:                code,
			ClientID:            clientID,
			UserID:              user.ID, // DB 의 UUID
			RedirectURI:         redirectURI,
			Scope:               scope,
			ExpiresAt:           time.Now().Add(10 * time.Minute),
			CodeChallenge:       codeChallenge,
			CodeChallengeMethod: codeChallengeMethod,
			Nonce:               nonce,
			AuthTime:            time.Now(),
		})
		AuditEvent(r, "login.success", "sub", user.ID, "client_id", clientID, "username", username)

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
