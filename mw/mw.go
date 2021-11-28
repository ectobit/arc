package mw

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.ectobit.com/lax"
)

// ZapLogger is middleware for logging requests using zap logger.
func ZapLogger(logger lax.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			start := time.Now()
			writer := middleware.NewWrapResponseWriter(res, req.ProtoMajor)

			next.ServeHTTP(writer, req)

			logger.Info(
				"request completed",
				lax.Time("time", start),
				lax.String("method", req.Method),
				lax.String("uri", req.RequestURI),
				lax.Int("status", writer.Status()),
				lax.Int("bytes", writer.BytesWritten()),
				lax.Duration("duration", time.Since(start)),
				lax.String("request_id", middleware.GetReqID(req.Context())))
		})
	}
}
