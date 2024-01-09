package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	issuer string = "rileythomp/jeopardy"

	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func SetJWTKeys() error {
	privateKeyStr := os.Getenv("JWT_RS512_KEY")
	privateKeyBytes := []byte(privateKeyStr)

	var err error
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		return err
	}

	publicKeyStr := os.Getenv("JWT_RS512_PUB_KEY")
	publicKeyBytes := []byte(publicKeyStr)

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return err
	}

	return nil
}

func GenerateJWT(id string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.MapClaims{
		"iss": issuer,
		"sub": id,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}).SignedString(privateKey)
}

func GetJWTSubject(jwtStr string) (string, error) {
	token, err := jwt.Parse(jwtStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return "", fmt.Errorf("Invalid signature")
		}
		return "", fmt.Errorf("Error parsing JWT: %s", err)
	}

	if !token.Valid {
		return "", fmt.Errorf("Invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("Error parsing claims")
	}

	iss, ok := claims["iss"].(string)
	if !ok {
		return "", fmt.Errorf("Error parsing issuer")
	}
	if iss != issuer {
		return "", fmt.Errorf("Invalid issuer")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", fmt.Errorf("Error parsing expiration")
	}
	if time.Now().Unix() > int64(exp) {
		return "", fmt.Errorf("Token expired")
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("Error parsing subject")
	}

	return sub, nil
}
