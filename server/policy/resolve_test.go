package policy

import (
	"testing"

	"github.com/ftery0/ouath/server/models"
)

// TestResolve 는 설계 문서 §4.3 진리표를 그대로 단위 테스트로 굳힌다.
// 핸들러를 띄우지 않고도 정책 분기의 회귀를 영구 차단한다.
func TestResolve(t *testing.T) {
	groupON := &models.ProjectGroup{ID: "marketing-tools", SSODefault: models.SSODefaultON}
	groupOFF := &models.ProjectGroup{ID: "admin-tools", SSODefault: models.SSODefaultOFF}

	clientIn := func(groupID string, override models.AppSSOOverride) *models.Client {
		return &models.Client{ClientID: "c-test", GroupID: groupID, SSOOverride: override}
	}
	clientNoGroup := &models.Client{ClientID: "c-no-group", GroupID: ""}

	cases := []struct {
		name           string
		hasSession     bool
		sessionGroupID string
		client         *models.Client
		group          *models.ProjectGroup
		prompt         string
		want           Decision
	}{
		// 1. 세션 없음, prompt 기본 → PROMPT
		{
			name:   "세션 없음, 기본 prompt",
			client: clientIn("marketing-tools", models.OverrideInherit),
			group:  groupON,
			want:   DecisionPrompt,
		},
		// 2. 세션 없음, prompt=none → ERROR (login_required)
		{
			name:   "세션 없음, prompt=none",
			prompt: "none",
			client: clientIn("marketing-tools", models.OverrideInherit),
			group:  groupON,
			want:   DecisionError,
		},
		// 3. 세션 있음, cross-group, prompt 기본 → PROMPT (Realm 경계)
		{
			name:           "cross-group, 기본 prompt",
			hasSession:     true,
			sessionGroupID: "admin-tools",
			client:         clientIn("marketing-tools", models.OverrideInherit),
			group:          groupON,
			want:           DecisionPrompt,
		},
		// 3b. 세션 있음, cross-group, prompt=none → ERROR
		{
			name:           "cross-group, prompt=none",
			hasSession:     true,
			sessionGroupID: "admin-tools",
			prompt:         "none",
			client:         clientIn("marketing-tools", models.OverrideInherit),
			group:          groupON,
			want:           DecisionError,
		},
		// 4. 같은 그룹 ON + INHERIT → SILENT ✨
		{
			name:           "같은 그룹 ON + INHERIT",
			hasSession:     true,
			sessionGroupID: "marketing-tools",
			client:         clientIn("marketing-tools", models.OverrideInherit),
			group:          groupON,
			want:           DecisionSilent,
		},
		// 5. 같은 그룹 ON + INHERIT + prompt=login → PROMPT (강제 재로그인)
		{
			name:           "같은 그룹 ON + prompt=login",
			hasSession:     true,
			sessionGroupID: "marketing-tools",
			prompt:         "login",
			client:         clientIn("marketing-tools", models.OverrideInherit),
			group:          groupON,
			want:           DecisionPrompt,
		},
		// 6. 같은 그룹 ON + FORCE_OFF → PROMPT (앱이 그룹 정책 거부)
		{
			name:           "같은 그룹 ON + FORCE_OFF",
			hasSession:     true,
			sessionGroupID: "marketing-tools",
			client:         clientIn("marketing-tools", models.OverrideForceOFF),
			group:          groupON,
			want:           DecisionPrompt,
		},
		// 7. 같은 그룹 OFF + INHERIT → PROMPT
		{
			name:           "같은 그룹 OFF + INHERIT",
			hasSession:     true,
			sessionGroupID: "admin-tools",
			client:         clientIn("admin-tools", models.OverrideInherit),
			group:          groupOFF,
			want:           DecisionPrompt,
		},
		// 8. 같은 그룹 OFF + FORCE_ON → SILENT (앱이 그룹 정책 무시)
		{
			name:           "같은 그룹 OFF + FORCE_ON",
			hasSession:     true,
			sessionGroupID: "admin-tools",
			client:         clientIn("admin-tools", models.OverrideForceON),
			group:          groupOFF,
			want:           DecisionSilent,
		},
		// 9. 같은 그룹 OFF + INHERIT + prompt=none → ERROR
		{
			name:           "같은 그룹 OFF + INHERIT + prompt=none",
			hasSession:     true,
			sessionGroupID: "admin-tools",
			prompt:         "none",
			client:         clientIn("admin-tools", models.OverrideInherit),
			group:          groupOFF,
			want:           DecisionError,
		},
		// 10. 세션 있음, client 가 그룹 미소속 → PROMPT (silent 영구 비활성)
		{
			name:           "세션 있음 + client 그룹 미소속",
			hasSession:     true,
			sessionGroupID: "marketing-tools",
			client:         clientNoGroup,
			group:          nil,
			want:           DecisionPrompt,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(Inputs{
				HasSession:     tc.hasSession,
				SessionGroupID: tc.sessionGroupID,
				Client:         tc.client,
				Group:          tc.group,
				Prompt:         tc.prompt,
			})
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
