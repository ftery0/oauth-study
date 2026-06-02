// Package policy 는 /oauth/authorize 진입 시 IdP 세션과 그룹/클라이언트 정책을
// 입력으로 받아 silent SSO / 로그인 폼 / 에러 중 어느 결정을 내릴지 계산한다.
//
// 핸들러에 정책 분기를 직접 박지 않고 별도 순수 함수로 분리하는 이유:
//   - 진리표 단위 테스트로 회귀를 영구 차단
//   - cross-group silent SSO 차단 같은 Realm 경계 invariant 를 한 곳에서 보장
//   - 추후 phase (PKCE, consent 등) 에서 같은 함수 시그니처에 입력만 더 추가하면 됨
package policy

import "github.com/ftery0/ouath/server/models"

// Decision: /oauth/authorize 처리 결과 분기.
type Decision string

const (
	// DecisionSilent: 로그인 폼 없이 즉시 auth code 발급 + redirect_uri 로 리다이렉트.
	DecisionSilent Decision = "SILENT"
	// DecisionPrompt: 로그인 폼 렌더링.
	DecisionPrompt Decision = "PROMPT"
	// DecisionError: silent 불가능한데 prompt=none 으로 비대화식 요청 → login_required 에러.
	DecisionError Decision = "ERROR"
)

// Inputs: Resolve 가 결정을 내리기 위한 모든 입력.
// 핸들러가 IdP 세션, 요청 client, 그룹, 쿼리스트링 prompt 를 채워 넘긴다.
type Inputs struct {
	HasSession     bool                 // IdP 세션 쿠키가 유효한가
	SessionGroupID string               // 세션의 LastGroupID (cross-group silent 차단 키)
	Client         *models.Client       // /oauth/authorize 요청한 client
	Group          *models.ProjectGroup // Client.GroupID 로 조회한 그룹 (없으면 nil)
	Prompt         string               // OIDC prompt: "", "none", "login"
}

// Resolve: 정책 결정 트리. 설계 문서 §4.2 의사코드와 동일.
//
// 결정 순서:
//  1. prompt=login → 무조건 PROMPT (세션 무시)
//  2. 세션 없음 + prompt=none → ERROR
//  3. 세션 없음 → PROMPT
//  4. 세션 있지만 client 가 그룹 미소속 → PROMPT (silent 영구 비활성)
//  5. 세션 있지만 다른 그룹 → PROMPT (또는 prompt=none 이면 ERROR) ← Realm 경계
//  6. 같은 그룹: effective = group.SSODefault, 단 FORCE_ON/FORCE_OFF 가 덮어씀
//     - effective == ON → SILENT
//     - effective == OFF → PROMPT (또는 prompt=none 이면 ERROR)
func Resolve(in Inputs) Decision {
	// 1) prompt=login 은 정책 무관 폼 강제
	if in.Prompt == "login" {
		return DecisionPrompt
	}

	// 2-3) 세션 없음
	if !in.HasSession {
		if in.Prompt == "none" {
			return DecisionError
		}
		return DecisionPrompt
	}

	// 4) Client 정보 없거나 그룹 미소속이면 silent 차단
	if in.Client == nil || in.Client.GroupID == "" {
		if in.Prompt == "none" {
			return DecisionError
		}
		return DecisionPrompt
	}

	// 5) Realm 경계: 세션이 다른 그룹에서 만들어졌으면 silent 금지
	if in.SessionGroupID != in.Client.GroupID {
		if in.Prompt == "none" {
			return DecisionError
		}
		return DecisionPrompt
	}

	// 6) 같은 그룹 — effective policy 계산
	var effective models.SSODefault
	if in.Group != nil {
		effective = in.Group.SSODefault
	}
	switch in.Client.SSOOverride {
	case models.OverrideForceON:
		effective = models.SSODefaultON
	case models.OverrideForceOFF:
		effective = models.SSODefaultOFF
	}

	if effective == models.SSODefaultON {
		return DecisionSilent
	}
	// effective == OFF (또는 그룹 정보 누락 시 안전상 OFF 로 처리)
	if in.Prompt == "none" {
		return DecisionError
	}
	return DecisionPrompt
}
