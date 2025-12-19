package middleware

import (
	"strings"

	"github.com/NSObjects/go-kit/code"
	"github.com/NSObjects/go-kit/errors"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
)

// JWTConfig holds JWT middleware configuration.
type JWTConfig struct {
	// SigningKey is the secret key for JWT validation.
	SigningKey []byte
	// SkipPaths are paths that skip JWT validation.
	SkipPaths []string
	// Enabled controls whether JWT is enabled.
	Enabled bool
	// ClaimsFunc creates a new claims instance.
	ClaimsFunc func(c echo.Context) jwt.Claims
}

// DefaultJWTConfig returns default JWT configuration.
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SigningKey: []byte("default-secret"),
		SkipPaths: []string{
			"/api/health",
			"/api/info",
			"/api/login",
		},
		Enabled: false,
	}
}

// JWT returns a JWT authentication middleware.
func JWT(config *JWTConfig) echo.MiddlewareFunc {
	if !config.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	cfg := echojwt.Config{
		SigningKey: config.SigningKey,
		Skipper: func(c echo.Context) bool {
			path := c.Path()
			for _, skipPath := range config.SkipPaths {
				// Support exact match and prefix match (with *)
				if path == skipPath {
					return true
				}
				if len(skipPath) > 0 && skipPath[len(skipPath)-1] == '*' &&
					strings.HasPrefix(path, skipPath[:len(skipPath)-1]) {
					return true
				}
			}
			return false
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return errors.WrapCode(err, code.ErrSignatureInvalid, "JWT signature invalid")
		},
	}

	if config.ClaimsFunc != nil {
		cfg.NewClaimsFunc = config.ClaimsFunc
	}

	return echojwt.WithConfig(cfg)
}

// CreateJWTConfig creates JWT config from parameters.
func CreateJWTConfig(secret string, skipPaths []string, enabled bool) *JWTConfig {
	return &JWTConfig{
		SigningKey: []byte(secret),
		SkipPaths:  skipPaths,
		Enabled:    enabled,
	}
}
