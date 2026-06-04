package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"github.com/ftery0/ouath/server/models"
	"github.com/ftery0/ouath/server/store"
)

// adminClientNewPageData: 신규 등록 폼에 넘기는 데이터.
// Phase-R: 그룹/override 제거. silent_sso 토글 기본값 true.
type adminClientNewPageData struct {
	ErrorMsg     string
	Name         string
	Description  string
	MainURL      string
	ServerURLs   []string
	RedirectURIs []string
	SilentSSO    bool
}

// adminClientCreatedPageData: 등록 성공 후 secret 1회 노출 페이지.
type adminClientCreatedPageData struct {
	Client *models.Client
}

// AdminClientNewFormHandler: GET /admin/clients/new — 등록 폼.
func AdminClientNewFormHandler(tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := adminClientNewPageData{
			RedirectURIs: []string{""},
			ServerURLs:   []string{""},
			SilentSSO:    true,
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
		silentSSO := r.FormValue("silent_sso") == "true"

		redirectURIs := cleanList(r.Form["redirect_uris"])
		serverURLs := cleanList(r.Form["server_urls"])

		if name == "" || len(redirectURIs) == 0 {
			data := adminClientNewPageData{
				Name:         name,
				Description:  description,
				MainURL:      mainURL,
				ServerURLs:   r.Form["server_urls"],
				RedirectURIs: r.Form["redirect_uris"],
				SilentSSO:    silentSSO,
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
			SilentSSO:    silentSSO,
		}

		if err := store.Clients.Register(c); err != nil {
			http.Error(w, "등록 실패: "+err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "admin_client_created.html", adminClientCreatedPageData{
			Client: c,
		})
	}
}

// AdminClientSilentSSOHandler: POST /admin/clients/{id}/silent-sso
// form: silent_sso=true|false
// JSON 응답.
func AdminClientSilentSSOHandler(w http.ResponseWriter, r *http.Request) {
	clientID := r.PathValue("id")
	if clientID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "id required"})
		return
	}
	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "form parse failed"})
		return
	}
	val := r.FormValue("silent_sso")
	if val != "true" && val != "false" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid silent_sso (must be 'true' or 'false')"})
		return
	}
	silentSSO := val == "true"

	if err := store.Clients.UpdateSilentSSO(clientID, silentSSO); err != nil {
		status := http.StatusNotFound
		if !errors.Is(err, errors.New("client not found")) {
			// 그냥 404 로 — 호출자가 ID 잘못 줬을 가능성이 큼
		}
		writeJSON(w, status, map[string]any{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"client_id":  clientID,
		"silent_sso": silentSSO,
	})
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
