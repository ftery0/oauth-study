package token

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"sub"`
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	jwt.RegisteredClaims
}

var (
	// 비대칭키 쌍: privateKey로 서명, PublicKey로 검증
	// 비대칭키를 쓰는 이유: 리소스 서버(Resource Server)가 비밀키 없이도 토큰을 검증할 수 있음
	// → 인가 서버만 서명 가능, 다른 서버들은 공개키로 검증만 가능 (보안상 안전)
	privateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey // JWKS 엔드포인트를 통해 외부에 공개
	issuerVal  string
)

func init() {
	var err error
	// 2048비트 RSA 키 생성 (프로덕션에서는 파일에서 로드해야 함)
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic("RSA 키 생성 실패: " + err.Error())
	}
	PublicKey = &privateKey.PublicKey
}

// Init: main에서 설정 로드 후 호출 (iss 클레임 설정)
func Init(_, issuer string) {
	issuerVal = issuer
}

func Create(userID, clientID, scope string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		ClientID: clientID,
		Scope:    scope,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    issuerVal,
		},
	}
	// HS512(대칭키) → RS256(비대칭키)로 변경
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
}

func Parse(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		// 서명 알고리즘이 RSA인지 반드시 확인 (알고리즘 혼동 공격 방지)
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("예상치 못한 서명 알고리즘: %v", t.Header["alg"])
		}
		return PublicKey, nil
	})
	if err != nil {
		return nil, err
	}
	return t.Claims.(*Claims), nil
}
