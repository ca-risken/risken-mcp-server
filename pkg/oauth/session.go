package oauth

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionData holds OAuth session information
type SessionData struct {
	State         string
	CodeChallenge string // Store code_challenge for PKCE verification
	RedirectURI   string
	ClientID      string
	CreatedAt     time.Time
}

// SessionManager interface for OAuth session management
type SessionManager interface {
	Store(data *SessionData) (string, error)
	Get(stateToken string) (*SessionData, bool)
}

// JWTSessionManager implements stateless JWT-based session management
type JWTSessionManager struct {
	signingKey []byte
	logger     *slog.Logger
}

// NewJWTSessionManager creates a new JWT-based session manager
func NewJWTSessionManager(signingKey []byte, logger *slog.Logger) *JWTSessionManager {
	return &JWTSessionManager{
		signingKey: signingKey,
		logger:     logger,
	}
}

// Store encodes session data into a JWT token and returns it as state
func (jsm *JWTSessionManager) Store(data *SessionData) (string, error) {
	claims := jwt.MapClaims{
		"state":          data.State,
		"code_challenge": data.CodeChallenge,
		"redirect_uri":   data.RedirectURI,
		"client_id":      data.ClientID,
		"iat":            time.Now().Unix(),
		"exp":            time.Now().Add(10 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jsm.signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}

	jsm.logger.Debug("Stored OAuth session in JWT",
		slog.String("original_state", data.State))

	return tokenString, nil
}

// Get decodes and validates JWT token to retrieve session data
func (jsm *JWTSessionManager) Get(stateToken string) (*SessionData, bool) {
	token, err := jwt.Parse(stateToken, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jsm.signingKey, nil
	})

	if err != nil {
		jsm.logger.Error("Failed to parse JWT session token", slog.String("error", err.Error()))
		return nil, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		jsm.logger.Error("Invalid JWT claims or token")
		return nil, false
	}

	// Extract claims
	state, _ := claims["state"].(string)
	codeChallenge, _ := claims["code_challenge"].(string)
	redirectURI, _ := claims["redirect_uri"].(string)
	clientID, _ := claims["client_id"].(string)

	if state == "" || codeChallenge == "" || redirectURI == "" {
		jsm.logger.Error("Missing required claims in JWT")
		return nil, false
	}

	sessionData := &SessionData{
		State:         state,
		CodeChallenge: codeChallenge,
		RedirectURI:   redirectURI,
		ClientID:      clientID,
		CreatedAt:     time.Now(), // Not needed for JWT but kept for interface compatibility
	}

	jsm.logger.Debug("Retrieved OAuth session from JWT",
		slog.String("original_state", state),
		slog.String("redirect_uri", redirectURI))

	return sessionData, true
}
