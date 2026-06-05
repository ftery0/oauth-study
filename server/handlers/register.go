package handlers

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// usernamePattern: 영문 소문자/숫자/언더스코어/하이픈, 3~32 자.
var usernamePattern = regexp.MustCompile(`^[a-z0-9_-]{3,32}$`)

// minPasswordLen: 학습 단계 최소 8 자. 운영에서는 zxcvbn 같은 강도 검증 권장.
const minPasswordLen = 8

// registerPageData: register.html 에 넘기는 데이터.
type registerPageData struct {
	ClientName          string
	ClientID            string
	RedirectURI         string
	State               string
	Scope               string
	Username            string // 에러 시 입력값 보존
	Email               string // 에러 시 입력값 보존
	ErrorMsg            string
	CSRFToken           string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
}

// emailPattern: 간단한 형식 검증. 학습 용으로 RFC 5322 까지는 안 감.
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// RegisterGetHandler: GET /oauth/register — 회원가입 폼.
// /oauth/authorize 와 같은 client_id / redirect_uri 검증을 거친 후 폼 표시.
func RegisterGetHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		clientID := q.Get("client_id")
		redirectURI := q.Get("redirect_uri")
		state := q.Get("state")
		scope := q.Get("scope")

		client, ok := store.Clients.GetByClientID(clientID)
		if !ok || !containsURI(client.RedirectURIs, redirectURI) {
			renderError(w, tmpl, "유효하지 않은 client_id 또는 redirect_uri 입니다")
			return
		}

		csrfToken, err := NewCSRFToken(w)
		if err != nil {
			http.Error(w, "csrf 토큰 생성 실패", http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "register.html", registerPageData{
			ClientName:          client.Name,
			ClientID:            clientID,
			RedirectURI:         redirectURI,
			State:               state,
			Scope:               scope,
			CSRFToken:           csrfToken,
			CodeChallenge:       q.Get("code_challenge"),
			CodeChallengeMethod: q.Get("code_challenge_method"),
			Nonce:               q.Get("nonce"),
		})
	}
}

// RegisterPostHandler: POST /oauth/register — 가입 처리 + 자동 로그인 + auth code 발급.
//
// 순서:
//  0. CSRF 검증
//  1. client_id / redirect_uri 재검증 (요청 변조 방어)
//  2. username / password 유효성
//  3. 중복 username 체크
//  4. bcrypt(password) → users INSERT
//  5. 세션 고정 방어 + IdP 세션 발급 + 쿠키 set
//  6. auth code 발급 → redirect_uri 로 안전 redirect
func RegisterPostHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 0. CSRF
		if err := VerifyCSRFToken(r); err != nil {
			http.Error(w, "CSRF 검증 실패", http.StatusForbidden)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "form parse 실패", http.StatusBadRequest)
			return
		}

		username := strings.TrimSpace(r.FormValue("username"))
		email := strings.TrimSpace(r.FormValue("email"))
		password := r.FormValue("password")
		passwordConfirm := r.FormValue("password_confirm")
		clientID := r.FormValue("client_id")
		redirectURI := r.FormValue("redirect_uri")
		state := r.FormValue("state")
		scope := r.FormValue("scope")

		// 1. client_id / redirect_uri 재검증
		client, ok := store.Clients.GetByClientID(clientID)
		if !ok || !containsURI(client.RedirectURIs, redirectURI) {
			renderError(w, tmpl, "유효하지 않은 client_id 또는 redirect_uri 입니다")
			return
		}

		// 폼 재렌더 헬퍼 (에러 메시지 + 입력값 보존)
		renderFormError := func(msg string) {
			csrfToken, _ := NewCSRFToken(w)
			tmpl.ExecuteTemplate(w, "register.html", registerPageData{
				ClientName:          client.Name,
				ClientID:            clientID,
				RedirectURI:         redirectURI,
				State:               state,
				Scope:               scope,
				Username:            username,
				Email:               email,
				ErrorMsg:            msg,
				CSRFToken:           csrfToken,
				CodeChallenge:       r.FormValue("code_challenge"),
				CodeChallengeMethod: r.FormValue("code_challenge_method"),
				Nonce:               r.FormValue("nonce"),
			})
		}

		// 2. username / email / password 유효성
		if !usernamePattern.MatchString(username) {
			renderFormError("아이디는 영문 소문자/숫자/_/-, 3~32 자")
			return
		}
		if email != "" && !emailPattern.MatchString(email) {
			renderFormError("이메일 형식이 올바르지 않습니다")
			return
		}
		if len(password) < minPasswordLen {
			renderFormError("비밀번호는 8 자 이상")
			return
		}
		if password != passwordConfirm {
			renderFormError("비밀번호가 일치하지 않습니다")
			return
		}

		// 3 + 4. INSERT (UNIQUE 충돌 시 ErrUserAlreadyExists)
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			http.Error(w, "비밀번호 hash 실패", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		newUser := &models.User{
			Username:     username,
			PasswordHash: string(hash),
			DisplayName:  username,
			Email:        email,
			// EmailVerified: 학습 범위 외 (실 운영에서는 메일 확인 후 true).
		}
		if err := store.Users.Create(ctx, newUser); err != nil {
			if errors.Is(err, store.ErrUserAlreadyExists) {
				renderFormError("이미 사용 중인 아이디입니다")
				return
			}
			http.Error(w, "회원가입 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 5. 세션 고정 방어 + 새 IdP 세션
		if oldSid, ok := GetIdPSessionID(r); ok {
			store.IdPSessions.Delete(oldSid)
		}
		sid, err := store.IdPSessions.Create(newUser.ID)
		if err != nil {
			http.Error(w, "세션 생성 실패", http.StatusInternalServerError)
			return
		}
		if err := SetIdPSessionCookie(w, sid); err != nil {
			http.Error(w, "세션 쿠키 설정 실패", http.StatusInternalServerError)
			return
		}
		ClearCSRFToken(w)

		// 6. auth code 발급 + 안전 redirect
		code, err := generateCode()
		if err != nil {
			http.Error(w, "서버 오류", http.StatusInternalServerError)
			return
		}
		store.AuthCodes.Store(code, models.AuthCode{
			Code:                code,
			ClientID:            clientID,
			UserID:              newUser.ID,
			RedirectURI:         redirectURI,
			Scope:               scope,
			ExpiresAt:           time.Now().Add(10 * time.Minute),
			CodeChallenge:       r.FormValue("code_challenge"),
			CodeChallengeMethod: r.FormValue("code_challenge_method"),
			Nonce:               r.FormValue("nonce"),
			AuthTime:            time.Now(),
		})
		AuditEvent(r, "register.success", "sub", newUser.ID, "client_id", clientID, "username", username)
		safeOAuthRedirect(w, r, redirectURI, map[string]string{
			"code":  code,
			"state": state,
		})
	}
}
