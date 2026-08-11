package dashboard

import (
	"testing"
	"time"
)

func TestSessionTokenLifecycle(t *testing.T) {
	tok := newSessionToken()
	if !validSessionToken(tok) {
		t.Fatal("fresh token should validate")
	}
	if validSessionToken("not-a-token") {
		t.Fatal("unknown token should not validate")
	}

	// backdate past the TTL → invalid, and pruned on the next login
	sessMu.Lock()
	tokens[tok] = time.Now().Add(-sessionTTL - time.Hour)
	sessMu.Unlock()
	if validSessionToken(tok) {
		t.Fatal("expired token should not validate")
	}

	_ = newSessionToken()
	sessMu.RLock()
	_, still := tokens[tok]
	sessMu.RUnlock()
	if still {
		t.Fatal("expired token should be pruned on new login")
	}
}
