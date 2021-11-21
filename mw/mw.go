package mw

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// ZapLogger is middleware for logging requests using zap logger.
func ZapLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			writer := middleware.NewWrapResponseWriter(res, req.ProtoMajor)

			next.ServeHTTP(writer, req)

			logger.Info(
				"request completed",
				zap.Time("time", start),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", writer.Status()),
				zap.Int("bytes", writer.BytesWritten()),
				zap.Duration("duration", time.Since(start)),
				zap.String("request_id", middleware.GetReqID(req.Context())))
		})
	}
}
