package handlers

import (
	"html/template"
	"net/http"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// adminMainPageData: admin_main.html 에 넘어가는 데이터.
type adminMainPageData struct {
	Groups   []*models.ProjectGroup
	Clients  []*models.Client
	FlashMsg string // ?created= / ?error= 같은 일회성 알림
	FlashErr bool
}

// AdminMainHandler: GET /admin — 등록된 service / group 목록.
func AdminMainHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flash, isErr := flashFromQuery(r)
		data := adminMainPageData{
			Groups:   store.Groups.All(),
			Clients:  store.Clients.All(),
			FlashMsg: flash,
			FlashErr: isErr,
		}
		tmpl.ExecuteTemplate(w, "admin_main.html", data)
	}
}

// flashFromQuery: 쿼리스트링 ?created= / ?error= 를 사용자 메시지로 변환.
func flashFromQuery(r *http.Request) (msg string, isErr bool) {
	if id := r.URL.Query().Get("created"); id != "" {
		return "그룹 \"" + id + "\" 가 등록되었습니다", false
	}
	switch r.URL.Query().Get("error") {
	case "invalid_group_id":
		return "그룹 ID 가 올바르지 않습니다 (영문 소문자로 시작, 3~32자, 소문자/숫자/하이픈만)", true
	case "name_required":
		return "그룹 이름이 필요합니다", true
	case "group_id_exists":
		return "이미 존재하는 그룹 ID 입니다", true
	case "register_failed":
		return "등록에 실패했습니다", true
	case "form_parse_failed":
		return "폼 파싱 실패", true
	}
	return "", false
}
