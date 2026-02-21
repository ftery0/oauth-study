package token

import (
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
	secretKey []byte
	issuerVal string
)

// Init: main에서 설정 로드 후 호출 (시크릿, iss 클레임 설정)
func Init(secret, issuer string) {
	secretKey = []byte(secret)
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
	return jwt.NewWithClaims(jwt.SigningMethodHS512, claims).SignedString(secretKey)
}

func Parse(tokenStr string) (*Claims, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}
	return t.Claims.(*Claims), nil
}
