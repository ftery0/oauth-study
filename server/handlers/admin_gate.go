package handlers

import (
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type adminLoginPageData struct {
	ErrorMsg string
}

// AdminLoginGetHandler: GET /admin/login — 폼 표시.
// 이미 인증되어 있으면 /admin 으로 보낸다.
func AdminLoginGetHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isAdminAuthed(r) {
			http.Redirect(w, r, "/admin", http.StatusFound)
			return
		}
		tmpl.ExecuteTemplate(w, "admin_login.html", adminLoginPageData{})
	}
}

// AdminLoginPostHandler: POST /admin/login — bcrypt 검증 + 쿠키 set.
func AdminLoginPostHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		password := r.FormValue("password")
		err := bcrypt.CompareHashAndPassword([]byte(adminPasswordHash), []byte(password))
		if err != nil {
			tmpl.ExecuteTemplate(w, "admin_login.html", adminLoginPageData{
				ErrorMsg: "비밀번호가 틀렸습니다",
			})
			return
		}
		if err := setAdminSessionCookie(w); err != nil {
			http.Error(w, "세션 쿠키 설정 실패", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/admin", http.StatusFound)
	}
}

// AdminLogoutHandler: POST /admin/logout — 쿠키 폐기.
func AdminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	clearAdminSessionCookie(w)
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// RequireAdminAuth: 미들웨어. 비인증이면 /admin/login 으로 리다이렉트.
func RequireAdminAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAdminAuthed(r) {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
