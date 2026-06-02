package models

import "time"

// AppSSOOverride: 앱 단위 SSO 정책 오버라이드.
// INHERIT 면 그룹의 SSODefault 를 따르고, FORCE_ON/FORCE_OFF 는
// 그룹 정책을 명시적으로 덮어쓴다.
type AppSSOOverride string

const (
	OverrideInherit  AppSSOOverride = "INHERIT"
	OverrideForceON  AppSSOOverride = "FORCE_ON"
	OverrideForceOFF AppSSOOverride = "FORCE_OFF"
)

// Client: OAuth 클라이언트 (서비스) 등록 정보
// 웹에서 서비스 등록 시 입력받는 필드 + OAuth 식별자
type Client struct {
	ID           string // 내부 UUID
	ClientID     string // OAuth client_id (노출용)
	ClientSecret string // OAuth client_secret (토큰 요청 시 검증)

	// 등록 폼에서 받는 정보
	Name         string   // 서비스명
	Description  string   // 서비스 설명
	MainURL      string   // 메인 URL
	ServerURLs   []string // 서버 URL 목록 (여러 개)
	RedirectURIs []string // 리다이렉트 URI 목록 (여러 개)

	OwnerID   string // 등록한 사용자 ID (웹 등록 시)
	CreatedAt time.Time

	// SSO 그룹 연결 (PR2 도입)
	GroupID     string         // 소속 그룹 ID. "" 면 그룹 미소속 → silent SSO 영구 비활성
	SSOOverride AppSSOOverride // 그룹 정책 오버라이드. 기본 OverrideInherit
}
