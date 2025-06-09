package oauth

import (
	"context"
	"fmt"
	"log/slog"
)

// Initialize loads JWKS and Authorization Server metadata from IdP
func (s *Server) Initialize(ctx context.Context) error {
	// Validate
	if err := validate.Struct(s.config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Session manager
	s.sessionManager = NewJWTSessionManager([]byte(s.config.JWTSigningKey), s.logger)

	// Load and validate authorization server metadata
	if err := s.LoadMetadata(ctx); err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	// Load JWKS for JWT validation
	if err := s.jwtValidator.LoadJWKS(ctx, s.oauth21Metadata); err != nil {
		return fmt.Errorf("failed to load JWKS: %w", err)
	}

	s.logger.Info("MCP Resource Server initialized successfully for Third-Party Authorization Flow",
		slog.String("idp_issuer", s.oauth21Metadata.Issuer),
		slog.String("authorization_endpoint", s.oauth21Metadata.AuthorizationEndpoint),
		slog.String("token_endpoint", s.oauth21Metadata.TokenEndpoint))

	return nil
}
