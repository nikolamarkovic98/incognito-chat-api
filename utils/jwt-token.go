package utils

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Username string
	jwt.StandardClaims
}

func GetJWTFromHeader(r *http.Request, chatId string) string {
	jwtToken := r.Header.Get("Authorization")
	if jwtToken == "" {
		return ""
	}

	tokenSplit := strings.Split(jwtToken, " ")
	if 2 > len(tokenSplit) {
		return ""
	}

	_, err := ParseJWTToken(tokenSplit[1], chatId)
	if err != nil {
		return ""
	}

	return jwtToken
}

func GenJWTToken(username string, chatId string) string {
	var jwtKey = []byte(os.Getenv("JWT_KEY") + chatId)
	expirationTime := time.Now().Add(time.Hour * 1)

	claims := Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "Error creating token"
	}

	return tokenString
}

func ParseJWTToken(token string, chatId string) (*jwt.Token, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &Claims{}, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid token")
		}

		var jwtKey = []byte(os.Getenv("JWT_KEY") + chatId)
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	return jwtToken, nil
}