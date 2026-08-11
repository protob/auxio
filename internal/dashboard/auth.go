package dashboard

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/protob/auxio/internal/config"
	"github.com/danielgtaylor/huma/v2"
)

type LoginInput struct {
	Body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
}

type LoginOutput struct {
	Body struct {
		Token string `json:"token"`
	}
}

func (h *Handler) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	// Both compares run before the branch - returning early on a bad username
	// would leak which of the two was wrong.
	userOK := subtle.ConstantTimeCompare([]byte(input.Body.Username), []byte(h.cfg.AdminUsername)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(input.Body.Password), []byte(h.cfg.AdminPassword)) == 1
	if !userOK || !passOK {
		return nil, huma.Error401Unauthorized("invalid credentials")
	}
	out := &LoginOutput{}
	out.Body.Token = newSessionToken()
	return out, nil
}

type contextKey string

const authCtxKey contextKey = "auth"

func DashboardAuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/health") {
				next.ServeHTTP(w, r)
				return
			}

			if strings.HasPrefix(r.URL.Path, "/api/openapi") {
				next.ServeHTTP(w, r)
				return
			}

			if r.URL.Path == "/api/login" {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, `{"error":"unauthorized","message":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			if strings.HasPrefix(auth, "Bearer ") {
				token := strings.TrimPrefix(auth, "Bearer ")
				if validSessionToken(token) {
					ctx := context.WithValue(r.Context(), authCtxKey, true)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			http.Error(w, `{"error":"unauthorized","message":"invalid credentials"}`, http.StatusUnauthorized)
		})
	}
}
