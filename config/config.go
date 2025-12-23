// Package config provides configuration management with hot-reload support.
package config

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Config is the main application configuration.
type Config struct {
	System   SystemConfig   `mapstructure:"system"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Mongodb  MongoConfig    `mapstructure:"mongodb"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
	Log      LogConfig      `mapstructure:"log"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	CORS     CORSConfig     `mapstructure:"cors"`
	Casbin   CasbinConfig   `mapstructure:"casbin"`
	Otel     OtelConfig     `mapstructure:"otel"`
}

// SystemConfig contains system-level settings.
type SystemConfig struct {
	Port    string `mapstructure:"port"`
	Env     string `mapstructure:"env"` // dev, test, prod
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Level   string `mapstructure:"level"`
}

// DatabaseConfig contains database connection settings.
// Supports mysql, postgres, and sqlite drivers.
type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`   // mysql, postgres, sqlite
	Host     string `mapstructure:"host"`     // mysql/postgres: server host
	Port     int    `mapstructure:"port"`     // mysql/postgres: server port
	User     string `mapstructure:"user"`     // mysql/postgres: username
	Password string `mapstructure:"password"` // mysql/postgres: password
	Database string `mapstructure:"database"` // database name or sqlite file path

	// MySQL specific
	Charset string `mapstructure:"charset"` // mysql: character set (default: utf8mb4)

	// PostgreSQL specific
	SSLMode string `mapstructure:"ssl_mode"` // postgres: disable, require, verify-ca, verify-full
	Schema  string `mapstructure:"schema"`   // postgres: search_path schema

	// Common options
	TimeZone        string `mapstructure:"timezone"`           // timezone for parsing time (e.g., "Asia/Shanghai", "Local")
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`     // max idle connections in pool
	MaxOpenConns    int    `mapstructure:"max_open_conns"`     // max open connections in pool
	MaxLifetime     int    `mapstructure:"max_lifetime"`       // max connection lifetime in seconds
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"` // max connection idle time in seconds
}

// RedisConfig contains Redis connection settings.
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"database"`
	Database int    `mapstructure:"database"` // Alias for DB
	PoolSize int    `mapstructure:"pool_size"`
}

// MongoConfig contains MongoDB connection settings.
type MongoConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

// KafkaConfig contains Kafka connection settings.
type KafkaConfig struct {
	Brokers  []string `mapstructure:"brokers"`
	ClientID string   `mapstructure:"client_id"`
	Topic    string   `mapstructure:"topic"`
}

// LogConfig contains logging settings.
type LogConfig struct {
	Level  string        `mapstructure:"level"`
	Format string        `mapstructure:"format"` // json, text, color
	Output string        `mapstructure:"output"` // stdout, stderr
	File   LogFileConfig `mapstructure:"file"`
}

// LogFileConfig contains file logging settings.
type LogFileConfig struct {
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	Format     string `mapstructure:"format"`
}

// JWTConfig contains JWT settings.
type JWTConfig struct {
	Secret    string        `mapstructure:"secret"`
	Expire    time.Duration `mapstructure:"expire"`
	SkipPaths []string      `mapstructure:"skip_paths"`
	Enabled   bool          `mapstructure:"enabled"`
}

// CORSConfig contains CORS settings.
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// CasbinConfig contains Casbin settings.
type CasbinConfig struct {
	Model      string   `mapstructure:"model"`
	ModelFile  string   `mapstructure:"model_file"`
	Enabled    bool     `mapstructure:"enabled"`
	SkipPaths  []string `mapstructure:"skip_paths"`
	AdminUsers []string `mapstructure:"admin_users"`
}

// OtelConfig contains OpenTelemetry settings.
type OtelConfig struct {
	Enabled       bool    `mapstructure:"enabled"`        // 是否启用 OpenTelemetry
	ServiceName   string  `mapstructure:"service_name"`   // 服务名称
	Environment   string  `mapstructure:"environment"`    // 环境标识: dev, test, prod
	OTLPEndpoint  string  `mapstructure:"otlp_endpoint"`  // OTLP gRPC 端点地址
	ExporterType  string  `mapstructure:"exporter_type"`  // 导出器类型: otlp, stdout
	SamplingRatio float64 `mapstructure:"sampling_ratio"` // 采样率: 0.0 - 1.0
}

// BaseConfig is an alias for Config for backward compatibility.
type BaseConfig = Config

// Source abstracts configuration loading from various sources.
type Source[T any] interface {
	Load(ctx context.Context) (T, error)
}

// WatchableSource supports hot-reload configuration.
type WatchableSource[T any] interface {
	Source[T]
	Watch(ctx context.Context, onChange func(T)) error
}

// Load loads configuration from the given path.
func Load[T any](path string) T {
	return LoadFrom(FileSource[T]{Path: path})
}

// LoadFrom loads configuration from a custom source.
func LoadFrom[T any](src Source[T]) T {
	c, err := src.Load(context.Background())
	if err != nil {
		panic(err)
	}
	return c
}

// FileSource loads configuration from a local file.
type FileSource[T any] struct {
	Path string
}

// Load loads configuration from file.
func (f FileSource[T]) Load(ctx context.Context) (T, error) {
	var zero T
	if f.Path == "" {
		return zero, nil
	}

	v := viper.New()

	// Set config type based on extension
	ext := ""
	if dot := strings.LastIndex(f.Path, "."); dot >= 0 {
		ext = strings.ToLower(f.Path[dot+1:])
	}
	switch ext {
	case "json":
		v.SetConfigType("json")
	case "yaml", "yml":
		v.SetConfigType("yaml")
	default:
		v.SetConfigType("toml")
	}

	// Read file
	content, err := os.ReadFile(f.Path)
	if err != nil {
		return zero, err
	}

	if err := v.ReadConfig(bytes.NewBuffer(content)); err != nil {
		return zero, err
	}

	// Support environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var c T
	if err := v.Unmarshal(&c); err != nil {
		return zero, err
	}

	return c, nil
}

// Watch watches for file changes and calls onChange.
func (f FileSource[T]) Watch(ctx context.Context, onChange func(T)) error {
	if f.Path == "" {
		return nil
	}

	v := viper.New()
	v.SetConfigFile(f.Path)
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		var c T
		if err := v.Unmarshal(&c); err == nil {
			onChange(c)
		}
	})

	return nil
}

// Store provides atomic read/update for configuration with hot-reload.
type Store[T any] struct {
	v    atomic.Value
	mu   sync.RWMutex
	subs map[string][]chan T
}

// NewStore creates a new configuration store.
func NewStore[T any](initial T) *Store[T] {
	s := &Store[T]{subs: make(map[string][]chan T)}
	s.v.Store(initial)
	return s
}

// Current returns the current configuration.
func (s *Store[T]) Current() T {
	c, _ := s.v.Load().(T)
	return c
}

// Update updates the configuration and notifies subscribers.
func (s *Store[T]) Update(c T) {
	s.v.Store(c)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.subs["*"] {
		select {
		case ch <- c:
		default:
		}
	}
}

// Subscribe subscribes to configuration updates.
func (s *Store[T]) Subscribe(key string) <-chan T {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan T, 1)
	s.subs[key] = append(s.subs[key], ch)
	return ch
}

// Bootstrap loads configuration from file and sets up hot-reload.
func Bootstrap[T any](path string) (T, *Store[T]) {
	cfg := Load[T](path)
	store := NewStore(cfg)

	// Set up file watching for hot-reload
	_ = FileSource[T]{Path: path}.Watch(context.Background(), func(newCfg T) {
		store.Update(newCfg)
	})

	return cfg, store
}

// NewCfg loads configuration from file (alias for Load).
func NewCfg[T any](path string) T {
	return Load[T](path)
}
