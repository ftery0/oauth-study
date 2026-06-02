package config

import (
	"log"
	"os"
)

// devDefaultJWTSecret: 개발 편의를 위한 fallback.
// production 환경에서는 이 값으로 동작하지 않도록 fail-fast 처리한다.
const devDefaultJWTSecret = "oauth-dev-secret-change-in-production"

type Config struct {
	Env       string // "development" | "production"
	Port      string
	Issuer    string
	JWTSecret string
}

func Load() Config {
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

	jwtSecret := os.Getenv("OAUTH_JWT_SECRET")
	if jwtSecret == "" {
		if env == "production" {
			log.Fatal("OAUTH_JWT_SECRET is required when APP_ENV=production")
		}
		jwtSecret = devDefaultJWTSecret
	} else if env == "production" && jwtSecret == devDefaultJWTSecret {
		log.Fatal("OAUTH_JWT_SECRET must not be the dev default in production")
	}

	return Config{
		Env:       env,
		Port:      port,
		Issuer:    issuer,
		JWTSecret: jwtSecret,
	}
}

// getEnv: 첫 번째 키 우선, 없으면 두 번째 키(하위호환) 사용
func getEnv(primary, fallback string) string {
	if v := os.Getenv(primary); v != "" {
		return v
	}
	return os.Getenv(fallback)
}
