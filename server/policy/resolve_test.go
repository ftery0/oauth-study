package policy

import (
	"testing"

	"github.com/ftery0/ouath/server/models"
)

// TestResolve: Phase-R 6 케이스 진리표.
func TestResolve(t *testing.T) {
	silentOn := &models.Client{ClientID: "c-on", SilentSSO: true}
	silentOff := &models.Client{ClientID: "c-off", SilentSSO: false}

	cases := []struct {
		name       string
		hasSession bool
		client     *models.Client
		prompt     string
		want       Decision
	}{
		{
			name:   "세션 없음, 기본 prompt",
			client: silentOn,
			want:   DecisionPrompt,
		},
		{
			name:   "세션 없음, prompt=none",
			prompt: "none",
			client: silentOn,
			want:   DecisionError,
		},
		{
			name:       "세션 있음 + silent_sso=true → SILENT",
			hasSession: true,
			client:     silentOn,
			want:       DecisionSilent,
		},
		{
			name:       "세션 있음 + silent_sso=true + prompt=login → PROMPT",
			hasSession: true,
			prompt:     "login",
			client:     silentOn,
			want:       DecisionPrompt,
		},
		{
			name:       "세션 있음 + silent_sso=false → PROMPT (매번 로그인)",
			hasSession: true,
			client:     silentOff,
			want:       DecisionPrompt,
		},
		{
			name:       "세션 있음 + silent_sso=false + prompt=none → ERROR",
			hasSession: true,
			prompt:     "none",
			client:     silentOff,
			want:       DecisionError,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Resolve(Inputs{
				HasSession: tc.hasSession,
				Client:     tc.client,
				Prompt:     tc.prompt,
			})
			if got != tc.want {
				t.Errorf("got %s, want %s", got, tc.want)
			}
		})
	}
}
