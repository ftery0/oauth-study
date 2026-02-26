package models

import "time"

// Client: OAuth 클라이언트 (서비스) 등록 정보
// 웹에서 서비스 등록 시 입력받는 필드 + OAuth 식별자
type Client struct {
	ID   string // 내부 UUID
	ClientID     string   // OAuth client_id (노출용)
	ClientSecret string   // OAuth client_secret (토큰 요청 시 검증)

	// 등록 폼에서 받는 정보
	Name         string   // 서비스명
	Description  string   // 서비스 설명
	MainURL      string   // 메인 URL
	ServerURLs   []string // 서버 URL 목록 (여러 개)
	RedirectURIs []string // 리다이렉트 URI 목록 (여러 개)

	OwnerID   string    // 등록한 사용자 ID (웹 등록 시)
	CreatedAt time.Time
}
