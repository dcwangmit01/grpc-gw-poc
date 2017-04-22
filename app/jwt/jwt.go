package jwt

import (
	"fmt"
	jwtgo "github.com/dgrijalva/jwt-go"
	"time"
)

var SigningKey = []byte("todo: randomize this key")

type CustomClaims struct {
	Id   string `json:"id"`
	Role string `json:"role"`
	jwtgo.StandardClaims
}

// Create the Claims

func CreateJwtWithIdRole(id string, role string) (string, error) {
	now := time.Now()
	claims := CustomClaims{
		id,
		role,
		jwtgo.StandardClaims{
			Issuer:    "apiservice",
			Audience:  "apiservice",
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(time.Minute * 60).Unix(),
		},
	}

	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS256, claims)
	ss, err := token.SignedString(SigningKey)
	fmt.Printf("%v %v", ss, err)

	return ss, err
}

func keyLookupFunction(token *jwtgo.Token) (interface{}, error) {
	// Always return the same SigningKey
	return SigningKey, nil
}

func ParseJwt(tokenStr string) (*jwtgo.Token, *CustomClaims, error) {
	token, err := jwtgo.ParseWithClaims(tokenStr, &CustomClaims{}, keyLookupFunction)

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		panic("Type Assertion failed")
	}
	return token, claims, err
}