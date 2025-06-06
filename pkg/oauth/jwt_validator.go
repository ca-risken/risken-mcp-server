package oauth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator validates JWT tokens from IdP
type JWTValidator struct {
	mcpServerURL    string
	oauth21Metadata *OAuth21Metadata
	keySet          *JWKSet
	logger          *slog.Logger
}

// JWKSet represents JSON Web Key Set
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// Claims represents JWT claims from IdP
type Claims struct {
	jwt.RegisteredClaims
	Scope    string   `json:"scope,omitempty"`
	Email    string   `json:"email,omitempty"`
	Name     string   `json:"name,omitempty"`
	Groups   []string `json:"groups,omitempty"`
	Username string   `json:"preferred_username,omitempty"`
}

// NewJWTValidator creates new JWT validator
func NewJWTValidator(mcpServerURL string, logger *slog.Logger) *JWTValidator {
	return &JWTValidator{
		mcpServerURL: mcpServerURL,
		logger:       logger,
	}
}

// LoadJWKS loads JSON Web Key Set from IdP
func (j *JWTValidator) LoadJWKS(ctx context.Context, metadata *OAuth21Metadata) error {
	j.oauth21Metadata = metadata
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, j.oauth21Metadata.JWKSURI, nil)
	if err != nil {
		return fmt.Errorf("failed to create JWKS request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("JWKS request failed with status: %d", resp.StatusCode)
	}

	var keySet JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&keySet); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	j.keySet = &keySet
	j.logger.Info("Loaded JWKS from IdP",
		slog.String("issuer", j.oauth21Metadata.Issuer),
		slog.Int("key_count", len(keySet.Keys)))

	return nil
}

// ValidateToken validates JWT access token from IdP
func (j *JWTValidator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify Signing Method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get Key ID
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Get Public Key
		publicKey, err := j.getPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Verify Issuer（IdP）
	if claims.Issuer != j.oauth21Metadata.Issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", j.oauth21Metadata.Issuer, claims.Issuer)
	}

	// Verify Expiration Time
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// getPublicKey retrieves RSA public key from JWKS
func (j *JWTValidator) getPublicKey(kid string) (*rsa.PublicKey, error) {
	if j.keySet == nil {
		return nil, fmt.Errorf("JWKS not loaded")
	}

	for _, key := range j.keySet.Keys {
		if key.Kid == kid && key.Kty == "RSA" {
			return j.jwkToRSAPublicKey(key)
		}
	}

	return nil, fmt.Errorf("key not found: %s", kid)
}

// jwkToRSAPublicKey converts JWK to RSA public key
func (j *JWTValidator) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode N parameter（modulus）
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N parameter: %w", err)
	}

	// Decode E parameter（exponent）
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E parameter: %w", err)
	}

	// Convert to big.Int
	n := new(big.Int).SetBytes(nBytes)
	e := 0
	for _, b := range eBytes {
		e = e<<8 + int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}
