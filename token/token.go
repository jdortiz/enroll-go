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
	"github.com/google/uuid"
)

const (
	sessionDuration         = time.Hour
	vonageExpirationSeconds = 600
)

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

// GetDecodingKey loads and parses the public key from the JWT_PUBLIC_KEY environment variable.
func GetDecodingKey() (crypto.PublicKey, error) {
	pemData := os.Getenv("JWT_PUBLIC_KEY")
	if pemData == "" {
		return nil, errors.New("JWT_PUBLIC_KEY environment variable not set")
	}

	// Support keys stored with literal \n in the env var (e.g. from .env files).
	pemData = strings.ReplaceAll(pemData, `\n`, "\n")

	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("JWT_PUBLIC_KEY does not contain a valid PEM block")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}

	return key, nil
}

// SessionClaims represents the payload of a JWT session token.
type SessionClaims struct {
	jwt.RegisteredClaims
}

// ProduceSessionToken generates a signed JWT for the given username using the provided private key.
func ProduceSessionToken(key crypto.PrivateKey, username string) (string, error) {
	method, err := signingMethodFor(key)
	if err != nil {
		return "", err
	}

	now := time.Now()
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(sessionDuration)),
			Subject:   username,
		},
	}

	return jwt.NewWithClaims(method, claims).SignedString(key)
}

// VonageClaims represents the payload of a Vonage JWT.
type VonageClaims struct {
	jwt.RegisteredClaims
	ApplicationID string `json:"application_id"`
}

// ProduceVonageToken generates a signed Vonage JWT using the provided private key.
func ProduceVonageToken(key crypto.PrivateKey) (string, error) {
	appID := os.Getenv("VONAGE_APP_ID")
	if appID == "" {
		return "", errors.New("VONAGE_APP_ID environment variable not set")
	}

	expiryDuration := vonageExpirationSeconds * time.Second
	now := time.Now()
	jti, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate jti: %w", err)
	}

	method, err := signingMethodFor(key)
	if err != nil {
		return "", err
	}

	claims := VonageClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiryDuration)),
			ID:        jti.String(),
		},
		ApplicationID: appID,
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
