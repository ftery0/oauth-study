package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// groupIDPattern: 그룹 ID 규칙. 영문 소문자/숫자/하이픈, 시작은 영문. 3~32자.
// 식별자처럼 안전한 글자만 허용해 URL/JSON 인코딩 이슈 회피.
var groupIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{2,31}$`)

// AdminGroupCreateHandler: POST /admin/groups
//
// 입력 (form):
//   - id          : "group-c" 같은 식별자 (영문 시작, 소문자/숫자/하이픈, 3~32)
//   - name        : 표시명
//   - description : 설명
//   - sso_default : "ON" | "OFF"
//
// 성공/실패 모두 /admin 으로 리다이렉트.
// 폼이 모달 안에 있어 별도 페이지 렌더링 안 함 — 에러는 ?error=... 로 표시.
func AdminGroupCreateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin?error=form_parse_failed", http.StatusFound)
		return
	}

	id := strings.TrimSpace(r.FormValue("id"))
	name := strings.TrimSpace(r.FormValue("name"))
	description := strings.TrimSpace(r.FormValue("description"))
	ssoDefault := r.FormValue("sso_default")

	if !groupIDPattern.MatchString(id) {
		http.Redirect(w, r, "/admin?error=invalid_group_id", http.StatusFound)
		return
	}
	if name == "" {
		http.Redirect(w, r, "/admin?error=name_required", http.StatusFound)
		return
	}
	if ssoDefault != string(models.SSODefaultON) && ssoDefault != string(models.SSODefaultOFF) {
		ssoDefault = string(models.SSODefaultOFF)
	}

	// 이미 존재하면 무시 (PG 의 ON CONFLICT DO NOTHING). 메모리는 덮어씀.
	if _, exists := store.Groups.Get(id); exists {
		http.Redirect(w, r, "/admin?error=group_id_exists", http.StatusFound)
		return
	}

	if err := store.Groups.Register(&models.ProjectGroup{
		ID:          id,
		Name:        name,
		Description: description,
		SSODefault:  models.SSODefault(ssoDefault),
	}); err != nil {
		http.Redirect(w, r, "/admin?error=register_failed", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/admin?created="+id, http.StatusFound)
}

// AdminGroupPolicyHandler: POST /admin/groups/{id}/policy
// form: sso_default=ON|OFF
// JSON 응답 — 어드민 모달의 토글 UI 가 즉시 반영하기 위해 reload 안 함.
func AdminGroupPolicyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id required"})
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "form parse failed"})
		return
	}
	ssoDefault := r.FormValue("sso_default")
	if ssoDefault != string(models.SSODefaultON) && ssoDefault != string(models.SSODefaultOFF) {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid sso_default"})
		return
	}
	if err := store.Groups.UpdateSSODefault(id, models.SSODefault(ssoDefault)); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":          true,
		"id":          id,
		"sso_default": ssoDefault,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
