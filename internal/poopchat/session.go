package poopchat

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	SessionLoggerKey = "SessionLogger"
	SessionIDKey     = "sessionID"
)

func SessionIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := uuid.NewString()
		ctx := context.WithValue(r.Context(), SessionLoggerKey, log.WithFields(log.Fields{
			SessionIDKey: sessionID,
		}))
		ctx = context.WithValue(ctx, SessionIDKey, sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
