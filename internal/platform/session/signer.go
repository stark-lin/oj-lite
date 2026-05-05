// Signs and verifies cookie payloads.

package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultCookieName   = "ojlite_session"
	DefaultSessionTTL   = 8 * time.Hour
	DefaultRefreshAfter = 2 * time.Hour
)

var (
	ErrSessionCookieMissing = errors.New("session cookie missing")
	ErrInvalidSession       = errors.New("invalid session")
	ErrSessionExpired       = errors.New("session expired")
)

type Manager struct {
	aead         cipher.AEAD
	randReader   io.Reader
	now          func() time.Time
	cookieName   string
	maxAge       time.Duration
	refreshAfter time.Duration
	path         string
	sameSite     http.SameSite
	secure       bool
}

func NewManager() (*Manager, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("generate session key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create aes-gcm: %w", err)
	}

	return &Manager{
		aead:         aead,
		randReader:   rand.Reader,
		now:          func() time.Time { return time.Now().UTC() },
		cookieName:   DefaultCookieName,
		maxAge:       DefaultSessionTTL,
		refreshAfter: DefaultRefreshAfter,
		path:         "/",
		sameSite:     http.SameSiteLaxMode,
	}, nil
}

func (manager *Manager) NewClaims(userID int64, role string, classroomID int64) Claims {
	now := manager.now()
	return Claims{
		UserID:      userID,
		Role:        role,
		ClassroomID: classroomID,
		IssuedAt:    now.Unix(),
		ExpiresAt:   now.Add(manager.maxAge).Unix(),
	}
}

func (manager *Manager) RefreshClaims(claims Claims) Claims {
	refreshed := claims
	now := manager.now()
	refreshed.IssuedAt = now.Unix()
	refreshed.ExpiresAt = now.Add(manager.maxAge).Unix()
	return refreshed
}

func (manager *Manager) ShouldRefresh(claims Claims) bool {
	return claims.Remaining(manager.now()) <= manager.refreshAfter
}

func (manager *Manager) Encode(claims Claims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal session claims: %w", err)
	}

	nonce := make([]byte, manager.aead.NonceSize())
	if _, err := io.ReadFull(manager.randReader, nonce); err != nil {
		return "", fmt.Errorf("generate session nonce: %w", err)
	}

	ciphertext := manager.aead.Seal(nil, nonce, payload, nil)
	token := append(nonce, ciphertext...)
	return base64.RawURLEncoding.EncodeToString(token), nil
}

func (manager *Manager) Decode(value string) (Claims, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return Claims{}, ErrInvalidSession
	}

	nonceSize := manager.aead.NonceSize()
	if len(raw) <= nonceSize {
		return Claims{}, ErrInvalidSession
	}

	nonce := raw[:nonceSize]
	ciphertext := raw[nonceSize:]
	payload, err := manager.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return Claims{}, ErrInvalidSession
	}

	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return Claims{}, ErrInvalidSession
	}

	if claims.UserID <= 0 || claims.Role == "" || claims.ExpiresAt <= claims.IssuedAt {
		return Claims{}, ErrInvalidSession
	}

	if claims.Expired(manager.now()) {
		return Claims{}, ErrSessionExpired
	}

	return claims, nil
}
