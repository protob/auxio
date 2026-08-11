package dashboard

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Sessions are in-memory only: a restart forces re-login, which is fine for a
// single operator. Persist them in the SQLite index if sticky sessions are
// ever wanted. Expired tokens are pruned on each new login, so nothing sweeps.
const sessionTTL = 30 * 24 * time.Hour

var (
	sessMu sync.RWMutex
	tokens = map[string]time.Time{} // token -> created
)

func newSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	t := hex.EncodeToString(b)
	now := time.Now()
	sessMu.Lock()
	for k, created := range tokens {
		if now.Sub(created) > sessionTTL {
			delete(tokens, k)
		}
	}
	tokens[t] = now
	sessMu.Unlock()
	return t
}

func validSessionToken(t string) bool {
	sessMu.RLock()
	created, ok := tokens[t]
	sessMu.RUnlock()
	return ok && time.Since(created) <= sessionTTL
}
