// Package token provides utilities for generating and managing JWT session tokens.
package token

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const sessionDuration = time.Hour

// GetEncodingKey loads and parses the private key from the JWT_SECRET environment variable.
func GetEncodingKey() (crypto.PrivateKey, error) {
	pemData := os.Getenv("JWT_SECRET")
	if pemData == "" {
		return nil, errors.New("JWT_SECRET environment variable not set")
	}

	// Support keys stored with literal \n in the env var (e.g. from .env files).
	pemData = strings.ReplaceAll(pemData, `\n`, "\n")

	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("JWT_SECRET does not contain a valid PEM block")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	return key.(crypto.PrivateKey), nil
}

// SessionClaims represents the payload of a JWT session token.
type SessionClaims struct {
	Sub string `json:"sub"`
	Exp int    `json:"exp"`
}

// ProduceSessionToken generates a signed JWT for the given username using the provided private key.
func ProduceSessionToken(key crypto.PrivateKey, username string) (string, error) {
	method, err := signingMethodFor(key)
	if err != nil {
		return "", err
	}

	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(sessionDuration).Unix(),
	}

	return jwt.NewWithClaims(method, claims).SignedString(key)
}

// signingMethodFor determines the appropriate JWT signing algorithm based on the private key type.
func signingMethodFor(key crypto.PrivateKey) (jwt.SigningMethod, error) {
	switch key.(type) {
	case *rsa.PrivateKey:
		return jwt.SigningMethodRS256, nil
	case *ecdsa.PrivateKey:
		return jwt.SigningMethodES256, nil
	case ed25519.PrivateKey:
		return jwt.SigningMethodEdDSA, nil
	default:
		return nil, errors.New("unsupported private key type")
	}
}
