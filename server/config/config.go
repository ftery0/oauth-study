package config

import (
	"os"
)

type Config struct {
	Port     string
	Issuer   string
	JWTSecret string
}

func Load() Config {
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
		jwtSecret = "oauth-dev-secret-change-in-production"
	}

	return Config{
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
