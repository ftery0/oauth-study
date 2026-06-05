package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"github.com/ftery0/ouath/server/config"
	"github.com/ftery0/ouath/server/db"
	"github.com/ftery0/ouath/server/handlers"
	"github.com/ftery0/ouath/server/router"
	"github.com/ftery0/ouath/server/store"
	pgstore "github.com/ftery0/ouath/server/store/postgres"
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
	handlers.IdPCookieInit(cfg.IdPSessionSecret)
	handlers.AdminInit(cfg.AdminPasswordHash, cfg.AdminSessionSecret)
	handlers.SetProduction(cfg.Env == "production")
	store.IdPSessions.StartCleanup()

	// Postgres 연결 + 마이그레이션 + store 교체 + 시드.
	if cfg.DatabaseURL != "" {
		ctx := context.Background()
		if err := db.Connect(ctx, cfg.DatabaseURL); err != nil {
			if cfg.Env == "production" {
				log.Fatal("DB 연결 실패: ", err)
			}
			log.Println("[dev] DB 연결 실패, 인메모리로 계속:", err)
		} else {
			if err := db.RunMigrations(ctx); err != nil {
				log.Fatal("migration 실패: ", err)
			}
			if err := db.BackfillClientSecretHashes(ctx); err != nil {
				log.Fatal("client_secret hash backfill 실패: ", err)
			}
			pgClients := pgstore.NewClientStore(db.Pool)
			pgUsers := pgstore.NewUserStore(db.Pool)
			// 시드 (이미 있으면 ON CONFLICT DO NOTHING / ErrUserAlreadyExists 흡수)
			if err := store.SeedClients(pgClients); err != nil {
				log.Fatal("seed clients: ", err)
			}
			if err := store.SeedUsers(ctx, pgUsers); err != nil {
				log.Fatal("seed users: ", err)
			}
			store.Clients = pgClients
			store.Users = pgUsers
			log.Println("Postgres 연결 + 시드 OK · store=postgres")
		}
	}

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
