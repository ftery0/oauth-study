package handlers

import (
	"html/template"
	"net/http"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// adminMainPageData: admin_main.html 에 넘어가는 데이터.
// Phase-R 단순화: Groups 제거. SilentSSOCount 추가.
type adminMainPageData struct {
	Clients        []*models.Client
	SilentSSOCount int
	FlashMsg       string
	FlashErr       bool
}

// AdminMainHandler: GET /admin — 등록된 service 목록.
func AdminMainHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flash, isErr := flashFromQuery(r)
		clients := store.Clients.All()
		silent := 0
		for _, c := range clients {
			if c.SilentSSO {
				silent++
			}
		}
		data := adminMainPageData{
			Clients:        clients,
			SilentSSOCount: silent,
			FlashMsg:       flash,
			FlashErr:       isErr,
		}
		tmpl.ExecuteTemplate(w, "admin_main.html", data)
	}
}

// flashFromQuery: 쿼리스트링 ?error= 를 사용자 메시지로 변환.
// Phase-R: 그룹 관련 메시지 제거.
func flashFromQuery(r *http.Request) (msg string, isErr bool) {
	switch r.URL.Query().Get("error") {
	case "register_failed":
		return "등록에 실패했습니다", true
	case "form_parse_failed":
		return "폼 파싱 실패", true
	}
	return "", false
}
