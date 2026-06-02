package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"net/http"
	"strings"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// adminClientNewPageData: 신규 등록 폼에 넘기는 데이터.
type adminClientNewPageData struct {
	Groups   []*models.ProjectGroup
	ErrorMsg string
	// 에러 발생 시 사용자가 입력한 값 보존
	Name         string
	Description  string
	MainURL      string
	ServerURLs   []string
	RedirectURIs []string
	GroupID      string
	SSOOverride  string
}

// adminClientCreatedPageData: 등록 성공 후 secret 1회 노출 페이지.
type adminClientCreatedPageData struct {
	Client *models.Client
}

// AdminClientNewFormHandler: GET /admin/clients/new — 등록 폼.
func AdminClientNewFormHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := adminClientNewPageData{
			Groups:       store.Groups.All(),
			SSOOverride:  string(models.OverrideInherit),
			RedirectURIs: []string{""},
			ServerURLs:   []string{""},
		}
		tmpl.ExecuteTemplate(w, "admin_client_new.html", data)
	}
}

// AdminClientCreateHandler: POST /admin/clients — 신규 client 등록.
func AdminClientCreateHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "form parse 실패", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		description := strings.TrimSpace(r.FormValue("description"))
		mainURL := strings.TrimSpace(r.FormValue("main_url"))
		groupID := r.FormValue("group_id")
		override := r.FormValue("sso_override")
		if override == "" {
			override = string(models.OverrideInherit)
		}

		// 배열 필드 — 빈 줄 제거
		redirectURIs := cleanList(r.Form["redirect_uris"])
		serverURLs := cleanList(r.Form["server_urls"])

		// 입력 검증
		if name == "" || len(redirectURIs) == 0 {
			data := adminClientNewPageData{
				Groups:       store.Groups.All(),
				Name:         name,
				Description:  description,
				MainURL:      mainURL,
				ServerURLs:   r.Form["server_urls"],
				RedirectURIs: r.Form["redirect_uris"],
				GroupID:      groupID,
				SSOOverride:  override,
				ErrorMsg:     "서비스명과 리다이렉트 URL 최소 1 개는 필수입니다",
			}
			if len(data.RedirectURIs) == 0 {
				data.RedirectURIs = []string{""}
			}
			if len(data.ServerURLs) == 0 {
				data.ServerURLs = []string{""}
			}
			tmpl.ExecuteTemplate(w, "admin_client_new.html", data)
			return
		}

		// client_id / client_secret 자동 발급
		clientID := "client-" + randomShort(8)
		clientSecret := randomShort(32)

		c := &models.Client{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Name:         name,
			Description:  description,
			MainURL:      mainURL,
			ServerURLs:   serverURLs,
			RedirectURIs: redirectURIs,
			OwnerID:      "",
			GroupID:      groupID,
			SSOOverride:  models.AppSSOOverride(override),
		}

		if err := store.Clients.Register(c); err != nil {
			http.Error(w, "등록 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// secret 은 이 순간 한 번만 노출 — 이후 다시 못 봄
		tmpl.ExecuteTemplate(w, "admin_client_created.html", adminClientCreatedPageData{
			Client: c,
		})
	}
}

// cleanList: 빈 줄 / 공백만인 줄 제거.
func cleanList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

// randomShort: hex 인코딩된 짧은 랜덤 (n 바이트).
func randomShort(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
