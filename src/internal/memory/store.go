package memory

import (
	"sync"
	"time"
)

// Session holds session token data
type Session struct {
	Email     string
	ExpiresAt time.Time
}

// In-memory token store
var (
	tokens   = make(map[string]Session)
	tokensMu sync.RWMutex
)

// SetToken Set Token stores a session token with expiration
func SetToken(token string, email string, duration time.Duration) {
	tokensMu.Lock()
	defer tokensMu.Unlock()

	tokens[token] = Session{
		Email:     email,
		ExpiresAt: time.Now().Add(duration),
	}
}

// GetToken Get Token checks if token exists and returns its session
func GetToken(token string) (Session, bool) {
	tokensMu.RLock()
	defer tokensMu.RUnlock()

	session, ok := tokens[token]
	if !ok || session.ExpiresAt.Before(time.Now()) {
		return Session{}, false
	}
	return session, true
}

// DeleteToken Delete Token removes a token (e.g. on logout)
func DeleteToken(token string) {
	tokensMu.Lock()
	defer tokensMu.Unlock()
	delete(tokens, token)
}

// ExtendToken Extend Token increases the expiration time of a valid token
func ExtendToken(token string, extra time.Duration) bool {
	tokensMu.Lock()
	defer tokensMu.Unlock()

	session, ok := tokens[token]
	if !ok || session.ExpiresAt.Before(time.Now()) {
		return false
	}

	session.ExpiresAt = session.ExpiresAt.Add(extra)
	tokens[token] = session

	return true
}

// ValidateToken Validate Token increases the expiration time of a valid token
func ValidateToken(token string) bool {
	tokensMu.Lock()
	defer tokensMu.Unlock()

	session, ok := tokens[token]
	if !ok || session.ExpiresAt.Before(time.Now()) {
		return false
	}

	return true
}
