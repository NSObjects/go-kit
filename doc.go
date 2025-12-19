/*
Package kit provides essential infrastructure components for Go API development.

go-kit is a collection of reusable packages for building production-ready
Go applications. It includes error handling, configuration management,
logging, middleware, and database connectivity.

# Core Packages

  - code: Error code framework with HTTP status mapping
  - config: Configuration loading with hot-reload support
  - log: Structured logging with multiple sinks
  - resp: Unified API response formatting
  - middleware: Echo middleware (JWT, Casbin, CORS, Recovery)
  - db: Database connection management (MySQL, Redis, MongoDB, Kafka)
  - health: Component health checking
  - cache: Redis cache abstraction
  - metrics: Prometheus metrics collection
  - utils: Common utilities
  - validator: Custom validation extensions
  - fx: Uber Fx module integrations

# Quick Start

	import (
	    "github.com/NSObjects/go-kit/code"
	    "github.com/NSObjects/go-kit/config"
	    "github.com/NSObjects/go-kit/log"
	)

	func main() {
	    cfg := config.Load("config.yaml")
	    logger := log.New(cfg.Log)

	    if err := doSomething(); err != nil {
	        return code.WrapDatabaseError(err, "operation failed")
	    }
	}

# Version

v1.0.0 - Stable release
*/
package kit
