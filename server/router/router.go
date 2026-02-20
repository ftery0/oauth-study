package router

import (
	"net/http"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()

	// Go 1.22부터 "METHOD /path" 형식으로 메서드별 라우팅 가능
	mux.HandleFunc("GET /oauth/authorize", stubHandler)  // 로그인 페이지 표시
	mux.HandleFunc("POST /oauth/login", stubHandler)     // 로그인 처리 → auth code 발급
	mux.HandleFunc("POST /oauth/token", stubHandler)     // auth code → access token 교환
	mux.HandleFunc("GET /oauth/userinfo", stubHandler)   // 유저 정보 반환 (Bearer 토큰 필요)

	return mux
}

// stubHandler: 아직 구현 전인 핸들러 임시 자리
func stubHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
