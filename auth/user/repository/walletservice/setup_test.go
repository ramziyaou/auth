package walletservice

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	ACCESS_SECRET = "testingaccess"
)

func GenerateTestToken(IIN string) (string, error) {
	accessTokenExp := time.Now().Add(20 * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":    false,
		"exp":      accessTokenExp,
		"iat":      1640024899,
		"iin":      IIN,
		"username": "sth",
		"userts":   "2021-12-31 19:36:36",
	})

	accessTokenString, err := accessToken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return "", err
	}

	return accessTokenString, nil
}
