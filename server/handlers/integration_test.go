package handlers

import (
	"html/template"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

// 통합 테스트는 AuthorizeHandler + LoginHandler 의 협력 흐름을 검증한다.
// 진리표 단위 테스트(policy/resolve_test.go) 와 달리, 실제 쿠키 발급/검증/
// 세션 흐름이 끝까지 동작하는지를 본다.

// ───── 테스트용 최소 템플릿 ─────
// 실제 login.html 마크업과 무관하게 핸들러가 채워주는 데이터만 노출하면 충분.
const testLoginTpl = `<form method="POST" action="/oauth/login">
<input name="csrf_token" value="{{.CSRFToken}}">
<input name="id"><input name="password">
<input name="state" value="{{.State}}">
<input name="client_id" value="{{.ClientID}}">
<input name="redirect_uri" value="{{.RedirectURI}}">
<input name="scope" value="{{.Scope}}">
{{if .ErrorMsg}}<div class="error">{{.ErrorMsg}}</div>{{end}}
</form>`

const testErrorTpl = `<div class="error">{{.ErrorMsg}}</div>`

// ───── 헬퍼 ─────

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	IdPCookieInit("integration-test-secret-32bytes!!")
	SetProduction(false)

	tmpl := template.New("")
	template.Must(tmpl.New("login.html").Parse(testLoginTpl))
	template.Must(tmpl.New("error.html").Parse(testErrorTpl))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /oauth/authorize", AuthorizeHandler(tmpl))
	mux.HandleFunc("POST /oauth/login", LoginHandler(tmpl))
	return httptest.NewServer(mux)
}

func newTestClient(t *testing.T) *http.Client {
	t.Helper()
	jar, _ := cookiejar.New(nil)
	return &http.Client{
		Jar: jar,
		// 302 를 따라가지 않고 응답 자체를 검사해야 함
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

var csrfRe = regexp.MustCompile(`name="csrf_token"\s+value="([^"]+)"`)

func extractCSRF(t *testing.T, body string) string {
	t.Helper()
	m := csrfRe.FindStringSubmatch(body)
	if len(m) < 2 {
		t.Fatalf("응답에 csrf_token 없음:\n%s", body)
	}
	return m[1]
}

func authorizeURL(srvURL string, extra url.Values) string {
	q := url.Values{
		"response_type": {"code"},
		"client_id":     {"app1"},
		"redirect_uri":  {"http://localhost:8011/callback"},
		"state":         {"t"},
		"scope":         {""},
	}
	for k, v := range extra {
		q[k] = v
	}
	return srvURL + "/oauth/authorize?" + q.Encode()
}

// loginViaForm: /authorize 응답에서 csrf 를 뽑아 /login 으로 POST.
// 성공 시 client 의 jar 에 idp_session 쿠키가 저장된다.
func loginViaForm(t *testing.T, srv *httptest.Server, client *http.Client) {
	t.Helper()

	resp, err := client.Get(authorizeURL(srv.URL, nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	csrf := extractCSRF(t, string(body))

	resp2, err := client.PostForm(srv.URL+"/oauth/login", url.Values{
		"id":           {"alice"},
		"password":     {"password123"},
		"client_id":    {"app1"},
		"redirect_uri": {"http://localhost:8011/callback"},
		"state":        {"t"},
		"scope":        {""},
		"csrf_token":   {csrf},
	})
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusFound {
		t.Fatalf("로그인 실패: status=%d", resp2.StatusCode)
	}
}

// ───── 시나리오 ─────

// CSRF 토큰 없이 POST /oauth/login → 403.
func TestIntegration_NoCSRF_PostBlocked(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	client := newTestClient(t)

	resp, err := client.PostForm(srv.URL+"/oauth/login", url.Values{
		"id":           {"alice"},
		"password":     {"password123"},
		"client_id":    {"app1"},
		"redirect_uri": {"http://localhost:8011/callback"},
		"state":        {"t"},
		"scope":        {""},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("got %d, want 403", resp.StatusCode)
	}
}

// 정상 로그인 흐름: /authorize → 200(폼+csrf+cookie) → POST → 302+code+idp_session.
func TestIntegration_FullLoginFlow_ReturnsCodeAndSetsSession(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	client := newTestClient(t)

	// 1) /authorize → 200, csrf 추출
	resp, err := client.Get(authorizeURL(srv.URL, nil))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	csrf := extractCSRF(t, string(body))

	// 2) POST /login → 302 + Location ?code=...&state=t
	resp2, err := client.PostForm(srv.URL+"/oauth/login", url.Values{
		"id":           {"alice"},
		"password":     {"password123"},
		"client_id":    {"app1"},
		"redirect_uri": {"http://localhost:8011/callback"},
		"state":        {"t"},
		"scope":        {""},
		"csrf_token":   {csrf},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusFound {
		t.Fatalf("got %d, want 302", resp2.StatusCode)
	}
	loc := resp2.Header.Get("Location")
	if !strings.HasPrefix(loc, "http://localhost:8011/callback?") || !strings.Contains(loc, "code=") {
		t.Errorf("Location 비정상: %s", loc)
	}

	// 3) idp_session 쿠키 set 확인
	var hasIdP bool
	for _, c := range resp2.Cookies() {
		if c.Name == "idp_session" && c.Value != "" && c.HttpOnly && c.SameSite == http.SameSiteLaxMode {
			hasIdP = true
			break
		}
	}
	if !hasIdP {
		t.Error("idp_session 쿠키 누락 또는 속성 오류 (HttpOnly + Lax 필요)")
	}
}

// silent SSO: 로그인 후 같은 client 로 /authorize → 폼 없이 302+code (Phase 1 핵심).
func TestIntegration_SilentSSO_WhenSessionMatchesGroup(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	client := newTestClient(t)

	loginViaForm(t, srv, client)

	// 다시 /authorize → silent 통과 기대
	resp, err := client.Get(authorizeURL(srv.URL, url.Values{"state": {"second"}}))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("silent 기대 302, got %d", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if !strings.Contains(loc, "code=") {
		t.Errorf("silent 응답에 code 누락: %s", loc)
	}
	if !strings.Contains(loc, "state=second") {
		t.Errorf("state 잘못됨: %s", loc)
	}
}

// Open Redirect 방어: state 인젝션이 querystring 으로 새지 않아야 함.
func TestIntegration_OpenRedirect_StateEscaped(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()
	client := newTestClient(t)

	// unsupported_response_type 으로 safeOAuthRedirect 경유시킴
	payload := "evil&error=injected&iss=evil.com"
	q := url.Values{
		"response_type": {"ATTACK"},
		"client_id":     {"app1"},
		"redirect_uri":  {"http://localhost:8011/callback"},
		"state":         {payload},
	}
	resp, err := client.Get(srv.URL + "/oauth/authorize?" + q.Encode())
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("got %d, want 302", resp.StatusCode)
	}

	u, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("location 파싱 실패: %v", err)
	}
	qq := u.Query()

	if got := qq["state"]; len(got) != 1 || got[0] != payload {
		t.Errorf("state 변형: got %v, want [%q]", got, payload)
	}
	// error 는 우리가 set 한 것 하나만 있어야 함 (인젝션된 error=injected 가 별도 키로 나오면 안 됨)
	if got := qq["error"]; len(got) != 1 || got[0] != "unsupported_response_type" {
		t.Errorf("error 키 오염: got %v", got)
	}
	// iss 같은 인젝션된 키가 등장하면 안 됨
	if qq.Get("iss") != "" {
		t.Errorf("iss 인젝션됨: %s", qq.Get("iss"))
	}
}
