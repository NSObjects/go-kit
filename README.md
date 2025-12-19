# go-kit

Essential infrastructure components for Go API development.

## Installation

```bash
go get github.com/NSObjects/go-kit
```

## Packages

| Package | Description |
|---------|-------------|
| `code` | Error code framework with HTTP status mapping |
| `config` | Configuration management with hot-reload |
| `log` | Structured logging with slog |
| `resp` | Unified API response formatting |
| `middleware` | Echo middleware (Error, Recovery, JWT, Casbin) |
| `db` | Database connections (MySQL, Redis, MongoDB) |
| `health` | Component health checking |
| `cache` | Redis cache abstraction |
| `metrics` | Prometheus metrics |
| `utils` | Common utilities |
| `validator` | Custom validation extensions |

## Quick Start

```go
package main

import (
    "github.com/NSObjects/go-kit/code"
    "github.com/NSObjects/go-kit/config"
    "github.com/NSObjects/go-kit/log"
    "github.com/NSObjects/go-kit/resp"
    "github.com/NSObjects/go-kit/middleware"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()
    e.HTTPErrorHandler = middleware.ErrorHandler
    e.Use(middleware.Recovery())

    e.GET("/users", listUsers)
    e.Logger.Fatal(e.Start(":8080"))
}

func listUsers(c echo.Context) error {
    users := []User{{ID: 1, Name: "Alice"}}
    return resp.ListDataResponse(c, users, 1)
}
```

## Version

v1.0.0
