// Package db provides database connection management.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/NSObjects/go-kit/config"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Manager manages all database connections.
type Manager struct {
	DB      *gorm.DB // Generic database connection (MySQL/PostgreSQL)
	Redis   *redis.Client
	MongoDB *mongo.Database
	Config  *config.BaseConfig
}

// NewManager creates a new database manager.
// ctx is used for connection timeouts during initialization.
func NewManager(ctx context.Context, cfg config.BaseConfig) (*Manager, error) {
	dm := &Manager{Config: &cfg}

	// Initialize database if configured
	if cfg.Database.Host != "" {
		db, err := NewDatabase(cfg.Database, os.Stdout)
		if err != nil {
			return nil, fmt.Errorf("database init: %w", err)
		}
		dm.DB = db
	}

	// Initialize Redis if configured
	if cfg.Redis.Host != "" {
		dm.Redis = NewRedis(cfg.Redis)
	}

	// Initialize MongoDB if configured
	if cfg.Mongodb.Host != "" {
		db, err := NewMongoDB(ctx, cfg.Mongodb)
		if err != nil {
			return nil, fmt.Errorf("mongodb init: %w", err)
		}
		dm.MongoDB = db
	}

	return dm, nil
}

// Start checks connectivity of all databases.
func (m *Manager) Start(ctx context.Context) error {
	// Check database connection
	if m.DB != nil {
		if sqlDB, err := m.DB.DB(); err == nil {
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("database ping: %w", err)
			}
		}
	}

	// Check Redis
	if m.Redis != nil {
		if err := m.Redis.Ping(ctx).Err(); err != nil {
			return fmt.Errorf("redis ping: %w", err)
		}
	}

	return nil
}

// Stop closes all database connections.
func (m *Manager) Stop(ctx context.Context) error {
	var errs []error

	// Close database connection
	if m.DB != nil {
		if sqlDB, err := m.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("database close: %w", err))
			}
		}
	}

	// Close Redis
	if m.Redis != nil {
		if err := m.Redis.Close(); err != nil {
			errs = append(errs, fmt.Errorf("redis close: %w", err))
		}
	}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// Health returns health status of all components.
func (m *Manager) Health(ctx context.Context) map[string]error {
	health := make(map[string]error)

	if m.DB != nil {
		if sqlDB, err := m.DB.DB(); err == nil {
			health["database"] = sqlDB.PingContext(ctx)
		} else {
			health["database"] = err
		}
	}

	if m.Redis != nil {
		health["redis"] = m.Redis.Ping(ctx).Err()
	}

	if m.MongoDB != nil {
		health["mongodb"] = nil // Simplified
	}

	return health
}

// DBWithContext returns the database connection with context.
func (m *Manager) DBWithContext(ctx context.Context) *gorm.DB {
	if m.DB == nil {
		return nil
	}
	return m.DB.WithContext(ctx)
}

// IsEnabled checks if a component is enabled.
func (m *Manager) IsEnabled(component string) bool {
	switch component {
	case "database", "mysql", "postgres":
		return m.DB != nil
	case "redis":
		return m.Redis != nil
	case "mongodb":
		return m.MongoDB != nil
	default:
		return false
	}
}

// NewDialector creates a GORM dialector based on the driver type.
func NewDialector(cfg config.DatabaseConfig) (gorm.Dialector, error) {
	switch cfg.Driver {
	case "mysql", "":
		charset := cfg.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		loc := cfg.TimeZone
		if loc == "" {
			loc = "Local"
		}
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, charset, loc)
		return mysql.Open(dsn), nil
	case "postgres":
		sslMode := cfg.SSLMode
		if sslMode == "" {
			sslMode = "disable"
		}
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, sslMode)
		if cfg.TimeZone != "" {
			dsn += fmt.Sprintf(" TimeZone=%s", cfg.TimeZone)
		}
		if cfg.Schema != "" {
			dsn += fmt.Sprintf(" search_path=%s", cfg.Schema)
		}
		return postgres.Open(dsn), nil
	case "sqlite":
		// SQLite uses Database field as file path (e.g., "test.db" or ":memory:")
		return sqlite.Open(cfg.Database), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// NewDatabase creates a database connection with connection pooling.
// logOutput is the writer for SQL logs (use io.Discard to suppress).
func NewDatabase(cfg config.DatabaseConfig, logOutput io.Writer) (*gorm.DB, error) {
	dialector, err := NewDialector(cfg)
	if err != nil {
		return nil, fmt.Errorf("create dialector: %w", err)
	}

	newLogger := logger.New(
		log.New(logOutput, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(dialector, &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Configure connection pool
	if sqlDB, err := db.DB(); err == nil {
		configurePool(sqlDB, cfg)
	}

	return db, nil
}

func configurePool(sqlDB *sql.DB, cfg config.DatabaseConfig) {
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.MaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	} else {
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}
	if cfg.ConnMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	}
}

// NewRedis creates a Redis client.
func NewRedis(cfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
}

// NewMongoDB creates a MongoDB connection.
// ctx is used for connection timeout.
func NewMongoDB(ctx context.Context, cfg config.MongoConfig) (*mongo.Database, error) {
	uri := "mongodb://"
	if cfg.User != "" && cfg.Password != "" {
		uri += cfg.User + ":" + cfg.Password + "@"
	}
	uri += fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return client.Database(cfg.Database), nil
}
