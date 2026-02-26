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

func LoginHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.FormValue("id")
		password := r.FormValue("password")
		state := r.FormValue("state")
		clientID := r.FormValue("client_id")
		redirectURI := r.FormValue("redirect_uri")
		scope := r.FormValue("scope")

		err := bcrypt.CompareHashAndPassword(
			[]byte(models.TestUser.PasswordHash),
			[]byte(password),
		)
		if id != models.TestUser.ID || err != nil {
			tmpl.ExecuteTemplate(w, "login.html", loginPageData{
				ClientName:  clientID,
				State:       state,
				ClientID:    clientID,
				RedirectURI: redirectURI,
				Scope:       scope,
				ErrorMsg:    "아이디 또는 비밀번호가 틀렸습니다",
			})
			return
		}

		// crypto/rand: 보안용 난수 생성 (math/rand는 예측 가능해서 사용 금지)
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

		http.Redirect(w, r,
			redirectURI+"?code="+code+"&state="+state,
			http.StatusFound,
		)
	}
}

func generateCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
