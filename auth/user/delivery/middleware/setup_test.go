package middleware

import (
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestMain(m *testing.M) {
	os.Setenv("ACCESS_SECRET", ACCESS_SECRET)
	os.Setenv("REFRESH_SECRET", REFRESH_SECRET)
	// os.Setenv("ACCESS_TTL", "20s")
	// os.Setenv("REFRESH_TTL", "10m")
	os.Exit(m.Run())
}

func GenerateToken() (string, error) {
	accessTokenExp := time.Now().Add(20 * time.Second).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin":     false,
		"exp":       accessTokenExp,
		"iat":       1640024899,
		"iin":       "910815450350",
		"username":  "sth",
		"createdAt": "2021-12-31 19:36:36",
	})

	accessTokenString, err := accessToken.SignedString([]byte(ACCESS_SECRET))

	if err != nil {
		return "", err
	}
	return accessTokenString, nil
}
