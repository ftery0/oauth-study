package models

import "time"

// SSODefault: 프로젝트 그룹의 기본 SSO 정책.
// 그룹에 속한 앱들이 별도 오버라이드를 두지 않으면 이 값을 따른다.
type SSODefault string

const (
	SSODefaultON  SSODefault = "ON"  // 그룹 멤버는 silent SSO 허용
	SSODefaultOFF SSODefault = "OFF" // 그룹 멤버라도 매번 로그인 필요
)

// ProjectGroup: 여러 클라이언트(앱) 를 묶는 SSO 경계 단위.
// 같은 그룹 내에서만 silent SSO 가 동작하고, 다른 그룹 앱에 가면
// IdP 세션이 있어도 로그인 폼이 다시 표시된다 (Realm 스타일).
type ProjectGroup struct {
	ID          string // "marketing-tools"
	Name        string // "마케팅 도구 그룹"
	Description string
	SSODefault  SSODefault
	CreatedAt   time.Time
}
