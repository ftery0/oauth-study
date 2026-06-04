// Package policy 는 /oauth/authorize 진입 시 IdP 세션과 client 의 silent_sso 토글
// 입력으로 받아 silent SSO / 로그인 폼 / 에러 중 어느 결정을 내릴지 계산한다.
//
// Phase-R 단순화: 그룹/Realm 모델 폐기. 글로벌 user pool + client 단위 토글.
//
// 입력 3 개로 6 케이스 진리표:
//
//	HasSession=false, prompt=""      → PROMPT
//	HasSession=false, prompt=none    → ERROR (login_required)
//	HasSession=true,  silent=true,  prompt=""    → SILENT ✨
//	HasSession=true,  silent=true,  prompt=login → PROMPT (강제 폼)
//	HasSession=true,  silent=false, prompt=""    → PROMPT
//	HasSession=true,  silent=false, prompt=none  → ERROR
package policy

import "github.com/ftery0/ouath/server/models"

// Decision: /oauth/authorize 처리 결과 분기.
type Decision string

const (
	DecisionSilent Decision = "SILENT" // 폼 없이 즉시 auth code 발급
	DecisionPrompt Decision = "PROMPT" // 로그인 폼 렌더링
	DecisionError  Decision = "ERROR"  // prompt=none 인데 silent 불가 → login_required
)

// Inputs: Resolve 가 결정을 내리기 위한 입력.
// Phase-R 단순화: SessionGroupID, Group 제거.
type Inputs struct {
	HasSession bool
	Client     *models.Client // .SilentSSO 만 본다
	Prompt     string         // "" / "none" / "login"
}

// Resolve: 정책 결정 트리.
func Resolve(in Inputs) Decision {
	// 1) prompt=login 은 정책 무관 폼 강제
	if in.Prompt == "login" {
		return DecisionPrompt
	}

	// 2) 세션 없음
	if !in.HasSession {
		if in.Prompt == "none" {
			return DecisionError
		}
		return DecisionPrompt
	}

	// 3) Client 정보 없거나 silent_sso=false → 폼
	if in.Client == nil || !in.Client.SilentSSO {
		if in.Prompt == "none" {
			return DecisionError
		}
		return DecisionPrompt
	}

	// 4) HasSession=true + Client.silent_sso=true + prompt 보통 → silent ✨
	return DecisionSilent
}
