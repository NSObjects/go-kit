package middleware

import (
	"github.com/NSObjects/go-kit/code"
	"github.com/NSObjects/go-kit/errors"
	"github.com/casbin/casbin/v2"
	casbin_mw "github.com/labstack/echo-contrib/casbin"
	"github.com/labstack/echo/v4"
)

// CasbinConfig holds Casbin middleware configuration.
type CasbinConfig struct {
	// Enabled controls whether Casbin is enabled.
	Enabled bool
	// SkipPaths are paths that skip authorization.
	SkipPaths []string
	// AdminUsers are users that bypass authorization.
	AdminUsers []string
	// UserGetter extracts user from context.
	UserGetter func(c echo.Context) (string, error)
	// EnforceHandler performs custom authorization logic.
	EnforceHandler func(c echo.Context, user string) (bool, error)
}

// DefaultCasbinConfig returns default Casbin configuration.
func DefaultCasbinConfig() *CasbinConfig {
	return &CasbinConfig{
		Enabled: false,
		SkipPaths: []string{
			"/api/health",
			"/api/info",
			"/api/login",
		},
		AdminUsers: []string{"root", "admin"},
	}
}

// Casbin returns an authorization middleware using Casbin.
func Casbin(enforcer *casbin.Enforcer, config *CasbinConfig) echo.MiddlewareFunc {
	if !config.Enabled || enforcer == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	cfg := casbin_mw.Config{
		Enforcer: enforcer,
		Skipper: func(c echo.Context) bool {
			path := c.Path()
			for _, skipPath := range config.SkipPaths {
				if path == skipPath {
					return true
				}
			}
			return false
		},
		ErrorHandler: func(c echo.Context, internal error, proposedStatus int) error {
			return errors.WrapCode(internal, code.ErrPermissionDenied, "permission denied")
		},
	}

	if config.UserGetter != nil {
		cfg.UserGetter = config.UserGetter
	}

	if config.EnforceHandler != nil {
		cfg.EnforceHandler = config.EnforceHandler
	}

	return casbin_mw.MiddlewareWithConfig(cfg)
}

// CreateCasbinConfig creates Casbin config from parameters.
func CreateCasbinConfig(enabled bool, skipPaths []string, adminUsers []string) *CasbinConfig {
	return &CasbinConfig{
		Enabled:    enabled,
		SkipPaths:  skipPaths,
		AdminUsers: adminUsers,
	}
}
