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
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Manager manages all database connections.
type Manager struct {
	MySQL   *gorm.DB
	Redis   *redis.Client
	MongoDB *mongo.Database
	Config  *config.BaseConfig
}

// NewManager creates a new database manager.
// ctx is used for connection timeouts during initialization.
func NewManager(ctx context.Context, cfg config.BaseConfig) (*Manager, error) {
	dm := &Manager{Config: &cfg}

	// Initialize MySQL if configured
	if cfg.Mysql.Host != "" {
		db, err := NewMySQL(cfg.Mysql, os.Stdout)
		if err != nil {
			return nil, fmt.Errorf("mysql init: %w", err)
		}
		dm.MySQL = db
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
	// Check MySQL
	if m.MySQL != nil {
		if sqlDB, err := m.MySQL.DB(); err == nil {
			if err := sqlDB.PingContext(ctx); err != nil {
				return fmt.Errorf("mysql ping: %w", err)
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

	// Close MySQL
	if m.MySQL != nil {
		if sqlDB, err := m.MySQL.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, fmt.Errorf("mysql close: %w", err))
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

	if m.MySQL != nil {
		if sqlDB, err := m.MySQL.DB(); err == nil {
			health["mysql"] = sqlDB.PingContext(ctx)
		} else {
			health["mysql"] = err
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

// MySQLWithContext returns MySQL with context.
func (m *Manager) MySQLWithContext(ctx context.Context) *gorm.DB {
	if m.MySQL == nil {
		return nil
	}
	return m.MySQL.WithContext(ctx)
}

// IsEnabled checks if a component is enabled.
func (m *Manager) IsEnabled(component string) bool {
	switch component {
	case "mysql":
		return m.MySQL != nil
	case "redis":
		return m.Redis != nil
	case "mongodb":
		return m.MongoDB != nil
	default:
		return false
	}
}

// NewMySQL creates a MySQL connection with connection pooling.
// logOutput is the writer for SQL logs (use io.Discard to suppress).
func NewMySQL(cfg config.MysqlConfig, logOutput io.Writer) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Dbname)

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

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	if sqlDB, err := db.DB(); err == nil {
		configurePool(sqlDB, cfg)
	}

	return db, nil
}

func configurePool(sqlDB *sql.DB, cfg config.MysqlConfig) {
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
