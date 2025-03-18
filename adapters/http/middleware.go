package http

import (
	"context"
	"net/http"

	"github.com/abhissng/neuron/utils/constant"
	"github.com/abhissng/neuron/utils/types"
	"github.com/google/uuid"
)

// Middleware adds a unique UUID to the context for each request.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		ctx := context.WithValue(r.Context(), types.RequestID(constant.RequestID), reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the UUID from the context.
func GetRequestID(ctx context.Context) string {
	reqID, ok := ctx.Value(constant.RequestID).(string)
	if !ok {
		return ""
	}
	return reqID
}
