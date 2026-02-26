package handlers

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/ftery0/ouath/server/token"
)

// JWK: JSON Web Key 형식 (RFC 7517)
type JWK struct {
	Kty string `json:"kty"` // Key Type: "RSA"
	Use string `json:"use"` // 용도: "sig" (서명용)
	Alg string `json:"alg"` // 알고리즘: "RS256"
	Kid string `json:"kid"` // Key ID: 키를 구분하는 식별자
	N   string `json:"n"`   // RSA 공개키 모듈러스 (base64url)
	E   string `json:"e"`   // RSA 공개키 지수 (base64url)
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWKSHandler: GET /oauth/jwks
// 공개키를 JWK Set 형식으로 반환
// 클라이언트는 이 엔드포인트에서 공개키를 가져와 JWT 서명을 직접 검증할 수 있음
// → 인가 서버에 매 요청마다 묻지 않아도 됨 (성능 향상)
func JWKSHandler(w http.ResponseWriter, r *http.Request) {
	pub := token.PublicKey

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: "ouath-key-1",
		// N: 모듈러스를 big-endian 바이트 배열 → base64url 인코딩
		N: base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		// E: 지수(보통 65537)를 big-endian 바이트 배열 → base64url 인코딩
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JWKS{Keys: []JWK{jwk}})
}
