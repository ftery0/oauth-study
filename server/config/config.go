package config

import (
	"log"
	"os"
)

// dev 기본 secret 들. production 환경에서는 이 값으로 동작하지 않도록 fail-fast.
const (
	devDefaultJWTSecret        = "oauth-dev-secret-change-in-production"
	devDefaultIdPSessionSecret = "idp-session-dev-secret-change-in-production"
)

type Config struct {
	Env              string // "development" | "production"
	Port             string
	Issuer           string
	JWTSecret        string
	IdPSessionSecret string // IdP 세션 쿠키 SecureCookie 키 도출용
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

	jwtSecret := requireSecret(env, "OAUTH_JWT_SECRET", devDefaultJWTSecret)
	idpSessionSecret := requireSecret(env, "IDP_SESSION_SECRET", devDefaultIdPSessionSecret)

	return Config{
		Env:              env,
		Port:             port,
		Issuer:           issuer,
		JWTSecret:        jwtSecret,
		IdPSessionSecret: idpSessionSecret,
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
