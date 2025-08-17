package memory

import (
	"sync"
	"time"
)

type Recovery struct {
	Token     string
	ExpiresAt time.Time
}

var (
	recoveries   = make(map[string]Recovery) // email -> token
	recoveriesMu sync.RWMutex
)

// SetRecoveryToken stores a recovery token for a user
func SetRecoveryToken(email, token string, duration time.Duration) {
	recoveriesMu.Lock()
	defer recoveriesMu.Unlock()
	recoveries[email] = Recovery{
		Token:     token,
		ExpiresAt: time.Now().Add(duration),
	}
}

// GetEmailByRecoveryToken finds email by token
func GetEmailByRecoveryToken(token string) (string, bool) {
	recoveriesMu.RLock()
	defer recoveriesMu.RUnlock()

	recovery, ok := recoveries[token]
	if !ok || recovery.ExpiresAt.Before(time.Now()) {
		return "", false
	}
	return recovery.Token, true

}

// DeleteRecoveryToken removes a recovery token
func DeleteRecoveryToken(email string) {
	recoveriesMu.Lock()
	defer recoveriesMu.Unlock()
	delete(recoveries, email)
}
