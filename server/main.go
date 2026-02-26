package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/ftery0/ouath/server/config"
	"github.com/ftery0/ouath/server/router"
	"github.com/ftery0/ouath/server/token"
)

// go:embed 지시어: 빌드 시 frontend 폴더 전체를 바이너리 안에 포함시킴
// 덕분에 배포할 때 HTML/CSS 파일을 별도로 챙길 필요가 없음
//
//go:embed frontend
var frontendFS embed.FS

func main() {
	cfg := config.Load()
	token.Init(cfg.JWTSecret, cfg.Issuer)

	// 템플릿 파싱: embed된 FS에서 templates/*.html 파일을 모두 읽어서 파싱
	tmpl, err := template.ParseFS(frontendFS, "frontend/templates/*.html")
	if err != nil {
		log.Fatal("템플릿 파싱 실패:", err)
	}

	// 정적 파일(CSS 등) 서빙: embed된 FS에서 static/ 경로를 HTTP로 노출
	staticFS, err := fs.Sub(frontendFS, "frontend/static")
	if err != nil {
		log.Fatal("static FS 생성 실패:", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux := router.New(tmpl)

	addr := ":" + cfg.Port
	fmt.Printf("oauth server running on %s\n", cfg.Issuer)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
