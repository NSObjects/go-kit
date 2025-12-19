package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Clock provides time operations (for testability).
type Clock interface {
	Now() time.Time
}

// RealClock is the default Clock implementation.
type RealClock struct{}

func (RealClock) Now() time.Time { return time.Now() }

// EncryptConfig holds password hashing configuration.
type EncryptConfig struct {
	// BcryptCost: recommended 10-14; default 12.
	BcryptCost int

	// EnablePrehash: SHA-256 the password before bcrypt to avoid 72-byte truncation.
	EnablePrehash bool

	// Pepper: server-side secret (optional). Use Base64 encoded environment variable.
	Pepper []byte
}

// DefaultEncryptConfig returns the default encryption configuration.
func DefaultEncryptConfig() EncryptConfig {
	return EncryptConfig{
		BcryptCost:    12,
		EnablePrehash: true,
		Pepper:        nil,
	}
}

// LoadPepperFromEnv loads pepper from environment variable.
func LoadPepperFromEnv(envKey string) ([]byte, error) {
	if envKey == "" {
		envKey = "PASS_PEPPER_B64"
	}
	v := strings.TrimSpace(os.Getenv(envKey))
	if v == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(v)
}

// PasswordHasher handles password hashing with dependency injection.
type PasswordHasher struct {
	config EncryptConfig
	clock  Clock
}

// NewPasswordHasher creates a new password hasher.
func NewPasswordHasher(config EncryptConfig) *PasswordHasher {
	return &PasswordHasher{
		config: config,
		clock:  RealClock{},
	}
}

// NewPasswordHasherWithClock creates a new password hasher with custom clock.
func NewPasswordHasherWithClock(config EncryptConfig, clock Clock) *PasswordHasher {
	return &PasswordHasher{
		config: config,
		clock:  clock,
	}
}

// Hash generates a bcrypt password hash.
func (h *PasswordHasher) Hash(plaintext string) (string, error) {
	var material []byte
	if h.config.EnablePrehash {
		material = prehash(plaintext, h.config.Pepper)
	} else {
		material = []byte(plaintext)
	}

	hash, err := bcrypt.GenerateFromPassword(material, h.config.BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify checks if plaintext matches stored hash.
// Returns: ok (match), needRehash (should upgrade cost), err.
func (h *PasswordHasher) Verify(storedHash, plaintext string) (ok bool, needRehash bool, err error) {
	if storedHash == "" {
		return false, false, errors.New("empty stored hash")
	}

	material := []byte(plaintext)
	if h.config.EnablePrehash {
		material = prehash(plaintext, h.config.Pepper)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), material); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, false, nil
		}
		return false, false, err
	}

	need, _ := h.NeedsRehash(storedHash)
	return true, need, nil
}

// NeedsRehash checks if a stored hash needs to be re-hashed with current config.
func (h *PasswordHasher) NeedsRehash(storedHash string) (bool, error) {
	cost, err := bcrypt.Cost([]byte(storedHash))
	if err != nil {
		return false, err
	}
	return cost < h.config.BcryptCost, nil
}

// RehashIfNeeded verifies password and re-hashes if needed.
func (h *PasswordHasher) RehashIfNeeded(storedHash, plaintext string) (newHash string, changed bool, err error) {
	ok, need, err := h.Verify(storedHash, plaintext)
	if err != nil {
		return "", false, err
	}
	if !ok {
		return "", false, errors.New("password mismatch")
	}
	if !need {
		return storedHash, false, nil
	}
	hash, err := h.Hash(plaintext)
	if err != nil {
		return "", false, err
	}
	return hash, true, nil
}

// TokenSafeNow returns current unix timestamp using injected clock.
func (h *PasswordHasher) TokenSafeNow() int64 {
	return h.clock.Now().Unix()
}

// prehash: SHA-256(password || 0x00 || pepper)
func prehash(password string, pepper []byte) []byte {
	h := sha256.New()
	h.Write([]byte(password))
	h.Write([]byte{0})
	if len(pepper) > 0 {
		h.Write(pepper)
	}
	return h.Sum(nil)
}
