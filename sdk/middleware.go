package ouath

import "net/http"

// RequireAuth: 로그인 안 된 사용자가 접근하면 /login으로 보냄
func (c *Client) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, ok := c.loadSession(r)
		if !ok || sess.UserID == "" {
			http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}
