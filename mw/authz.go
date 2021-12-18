package mw

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/casbin/casbin"
	"github.com/go-chi/jwtauth/v5"
	"go.ectobit.com/lax"
)

// ErrInvalidSubject is returned when there is no sub within JWT claim or when it is not of a string type.
var ErrInvalidSubject = errors.New("invalid subject")

// Authorizer is middleware to enforce Casbin authorization.
func Authorizer(enforcer *casbin.Enforcer, log lax.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			user, err := SubjectFromJWT(req.Context())
			if err != nil {
				log.Warn("authorizer", lax.Error(err))
				http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

				return
			}

			if !enforcer.Enforce(user, req.URL.Path, req.Method) {
				http.Error(res, http.StatusText(http.StatusForbidden), http.StatusForbidden)

				return
			}

			next.ServeHTTP(res, req)
		})
	}
}

// SubjectFromJWT find out subject claim from context.
func SubjectFromJWT(ctx context.Context) (string, error) {
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("claims from jwt: %w", err)
	}

	sub, ok := claims["sub"] //nolint:varnamelen
	if !ok {
		return "", fmt.Errorf("%w: not found", ErrInvalidSubject)
	}

	s, ok := sub.(string)
	if !ok {
		return "", fmt.Errorf("%w: not string type", ErrInvalidSubject)
	}

	return s, nil
}
