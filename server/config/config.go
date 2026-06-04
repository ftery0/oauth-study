package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// loadDotEnvOnce: cwd 의 .env 와 server 디렉토리의 .env 를 둘 다 시도한다.
// 이미 process env 에 설정된 값은 덮어쓰지 않는다 (env 변수 우선).
func loadDotEnvOnce() {
	_ = godotenv.Load(".env")
	if exe, err := os.Executable(); err == nil {
		_ = godotenv.Load(filepath.Join(filepath.Dir(exe), ".env"))
	}
}

// dev 기본 secret 들. production 환경에서는 이 값으로 동작하지 않도록 fail-fast.
const (
	devDefaultJWTSecret          = "oauth-dev-secret-change-in-production"
	devDefaultIdPSessionSecret   = "idp-session-dev-secret-change-in-production"
	devDefaultAdminSessionSecret = "admin-session-dev-secret-change-in-production"
	// devDefaultAdminPasswordHash: bcrypt("admin", cost=12) 학습용. production 에서는 반드시 교체.
	// 새 hash 생성: go run ./cmd/hashgen <password>
	devDefaultAdminPasswordHash = "$2a$12$M5kxy0ym964CE65bZG2vt.HwHi1Wt/FnGoIvVsI//cXSMFKevEWAO"
)

type Config struct {
	Env                string // "development" | "production"
	Port               string
	Issuer             string
	JWTSecret          string
	IdPSessionSecret   string // IdP 세션 쿠키 SecureCookie 키 도출용
	AdminPasswordHash  string // bcrypt hash. 어드민 게이트 비밀번호 검증
	AdminSessionSecret string // 어드민 세션 쿠키 SecureCookie 키 도출용
	DatabaseURL        string // Postgres DSN. 비어있으면 DB 연결 시도하지 않음 (P2-B)
}

func Load() Config {
	loadDotEnvOnce()

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	port := getEnv("OAUTH_PORT", "OUATH_PORT")
	if port == "" {
		port = "8080"
	}

	issuer := getEnv("OAUTH_ISSUER", "OUATH_ISSUER")
	if issuer == "" {
		issuer = "http://localhost:" + port
	}

	jwtSecret := requireSecret(env, "OAUTH_JWT_SECRET", devDefaultJWTSecret)
	idpSessionSecret := requireSecret(env, "IDP_SESSION_SECRET", devDefaultIdPSessionSecret)
	adminSessionSecret := requireSecret(env, "OAUTH_ADMIN_SESSION_SECRET", devDefaultAdminSessionSecret)
	adminPasswordHash := requireSecret(env, "OAUTH_ADMIN_PASSWORD_HASH", devDefaultAdminPasswordHash)

	// DATABASE_URL: 비어있으면 DB 연결 안 함 (P2-B 학습 호환).
	// 권장 형식: postgres://oauth:oauth-dev@localhost:5433/oauth?sslmode=disable
	databaseURL := os.Getenv("DATABASE_URL")

	return Config{
		Env:                env,
		Port:               port,
		Issuer:             issuer,
		JWTSecret:          jwtSecret,
		IdPSessionSecret:   idpSessionSecret,
		AdminPasswordHash:  adminPasswordHash,
		AdminSessionSecret: adminSessionSecret,
		DatabaseURL:        databaseURL,
	}
}

// requireSecret: env 에서 secret 을 읽되, production 에서는 빈 값/dev 기본값을 모두 거부한다.
// 한 곳에 모아 관리해서 secret 마다 fail-fast 로직이 흩어지지 않도록 한다.
func requireSecret(env, envKey, devDefault string) string {
	v := os.Getenv(envKey)
	if v == "" {
		if env == "production" {
			log.Fatalf("%s is required when APP_ENV=production", envKey)
		}
		return devDefault
	}
	if env == "production" && v == devDefault {
		log.Fatalf("%s must not be the dev default in production", envKey)
	}
	return v
}

// getEnv: 첫 번째 키 우선, 없으면 두 번째 키(하위호환) 사용
func getEnv(primary, fallback string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(fallback)
}
