package ouath

import (
	"crypto/sha256"
	"net/http"
)

type Config struct {
	ServerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	Scopes        []string
	SessionSecret string // 세션 쿠키 서명/암호화용 (32자 이상 권장)
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func New(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{},
	}
}

// deriveKeys: SessionSecret에서 hashKey(32), blockKey(32) 생성
func deriveKeys(secret string) (hashKey, blockKey []byte) {
	h := sha256.Sum256([]byte(secret))
	hashKey = h[:]
	b := sha256.Sum256([]byte(secret + ".enc"))
	blockKey = b[:]
	return
}
