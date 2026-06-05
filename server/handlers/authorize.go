package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/policy"
	"github.com/ftery0/ouath/server/store"
)

// loginPageData: login.html 템플릿에 넘겨줄 데이터 구조체.
// CSRFToken 은 PR4b 에서 Double-Submit Cookie 용으로 추가.
// PKCE/Nonce 는 hidden 필드로 흘려 보내 로그인 후 발급되는 AuthCode 에 동승.
type loginPageData struct {
	ClientName          string
	State               string
	ClientID            string
	RedirectURI         string
	Scope               string
	ErrorMsg            string
	CSRFToken           string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
}

// AuthorizeHandler: GET /oauth/authorize 처리.
//
// 흐름:
//  1. redirect_uri / client_id / response_type 검증 (Open Redirect 방어 — 기존 유지)
//  2. IdP 세션 + 그룹 조회
//  3. policy.Resolve → SILENT / PROMPT / ERROR 분기
//
// SILENT 분기에서 폼 없이 즉시 auth code 가 발급되는 것이 silent SSO 의 본질.
func AuthorizeHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		responseType := q.Get("response_type")
		clientID := q.Get("client_id")
		redirectURI := q.Get("redirect_uri")
		state := q.Get("state")
		scope := q.Get("scope")
		prompt := q.Get("prompt")
		codeChallenge := q.Get("code_challenge")
		codeChallengeMethod := q.Get("code_challenge_method")
		nonce := q.Get("nonce")

		// 1. redirect_uri 검증 — 가장 먼저 (Open Redirect 방어)
		client, ok := store.Clients.GetByClientID(clientID)
		if !ok || !containsURI(client.RedirectURIs, redirectURI) {
			renderError(w, tmpl, "유효하지 않은 client_id 또는 redirect_uri입니다")
			return
		}

		// 2. response_type 검증 — 에러도 redirect_uri 로 안전 반환
		if responseType != "code" {
			safeOAuthRedirect(w, r, redirectURI, map[string]string{
				"error": "unsupported_response_type",
				"state": state,
			})
			return
		}

		// 2-b. PKCE method 검증 — challenge 가 있으면 S256 만 허용 (OAuth 2.1).
		if codeChallenge != "" {
			if codeChallengeMethod == "" {
				codeChallengeMethod = "plain"
			}
			if codeChallengeMethod != "S256" {
				safeOAuthRedirect(w, r, redirectURI, map[string]string{
					"error":             "invalid_request",
					"error_description": "code_challenge_method must be S256",
					"state":             state,
				})
				return
			}
		}

		// 3. IdP 세션 조회 — Phase-R 단순화: LastGroupID 더 이상 안 본다
		var (
			hasSession bool
			userID     string
		)
		if sid, ok := GetIdPSessionID(r); ok {
			if sess, ok := store.IdPSessions.Get(sid); ok {
				hasSession = true
				userID = sess.UserID
			}
		}

		// 4. 정책 결정 — 그룹 입력 제거 (HasSession, Client.SilentSSO, Prompt 3 입력)
		decision := policy.Resolve(policy.Inputs{
			HasSession: hasSession,
			Client:     client,
			Prompt:     prompt,
		})

		switch decision {
		case policy.DecisionSilent:
			// 폼 없이 즉시 auth code 발급 → redirect_uri 로 반환
			code, err := generateCode()
			if err != nil {
				http.Error(w, "서버 오류", http.StatusInternalServerError)
				return
			}
			store.AuthCodes.Store(code, models.AuthCode{
				Code:                code,
				ClientID:            clientID,
				UserID:              userID,
				RedirectURI:         redirectURI,
				Scope:               scope,
				ExpiresAt:           time.Now().Add(10 * time.Minute),
				CodeChallenge:       codeChallenge,
				CodeChallengeMethod: codeChallengeMethod,
				Nonce:               nonce,
				AuthTime:            time.Now(),
			})
			safeOAuthRedirect(w, r, redirectURI, map[string]string{
				"code":  code,
				"state": state,
			})

		case policy.DecisionError:
			// prompt=none 인데 silent 불가 → OIDC 표준 login_required
			safeOAuthRedirect(w, r, redirectURI, map[string]string{
				"error": "login_required",
				"state": state,
			})

		case policy.DecisionPrompt:
			// 로그인 폼 렌더 + Double-Submit CSRF 토큰 발급
			csrfToken, err := NewCSRFToken(w)
			if err != nil {
				http.Error(w, "csrf 토큰 생성 실패", http.StatusInternalServerError)
				return
			}
			data := loginPageData{
				ClientName:          client.Name,
				State:               state,
				ClientID:            clientID,
				RedirectURI:         redirectURI,
				Scope:               scope,
				CSRFToken:           csrfToken,
				CodeChallenge:       codeChallenge,
				CodeChallengeMethod: codeChallengeMethod,
				Nonce:               nonce,
			}
			if err := tmpl.ExecuteTemplate(w, "login.html", data); err != nil {
				http.Error(w, "템플릿 렌더링 실패", http.StatusInternalServerError)
			}
		}
	}
}

// renderError: redirect_uri 자체가 잘못된 경우 전용 에러 페이지 표시.
// 이때는 redirect 하면 안 되므로 에러 페이지를 직접 렌더링한다.
func renderError(w http.ResponseWriter, tmpl *template.Template, msg string) {
	w.WriteHeader(http.StatusBadRequest)
	tmpl.ExecuteTemplate(w, "error.html", loginPageData{ErrorMsg: msg})
}

// containsURI: 허용된 URI 목록에 포함되어 있는지 검사.
// 반드시 완전 일치(exact match) 로 비교해야 한다.
// 접두사 일치를 쓰면 공격자가 example.com.evil.com 같은 걸 통과시킬 수 있다.
func containsURI(uris []string, target string) bool {
	for _, u := range uris {
		if u == target {
			return true
		}
	}
	return false
}
