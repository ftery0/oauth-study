package config

import (
	"os"
)

type Config struct {
	Port   string
	Issuer string
}

func Load() Config {
	port := os.Getenv("OUATH_PORT")
	if port == "" {
		port = "8080"
	}

	issuer := os.Getenv("OUATH_ISSUER")
	if issuer == "" {
		// Issuer는 JWT의 iss 클레임에 들어가는 서버 식별자
		issuer = "http://localhost:" + port
	}

	return Config{
		Port:   port,
		Issuer: issuer,
	}
}
