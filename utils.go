package main

import (
	"github.com/dgrijalva/jwt-go"
	"log"
	"math/rand"
	"time"
)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Token struct {
	AccessToken           string `json:"access_token"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int    `json:"expires_in"`
	Scope                 string `json:"scope"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	IdToken               string `json:"id_token"`
	SessionState          string `json:"session_state"`
}

func (z *Token) DecodeIdToken() jwt.MapClaims {
	token, _, err := new(jwt.Parser).ParseUnverified(z.IdToken, jwt.MapClaims{})
	if err != nil {
		log.Println("Failed to parse token:", err)
		return nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		log.Println("Failed to parse claims")
		return nil
	}
	return claims
}

func GetSubFromIdToken(claims jwt.MapClaims) string {
	return claims["sub"].(string)
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	rand.Seed(time.Now().UnixNano())

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func randomStr(length int) string {
	return stringWithCharset(length, charset)
}
