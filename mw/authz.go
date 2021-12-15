package mw

import (
	"net/http"

	"github.com/casbin/casbin"
	"github.com/go-chi/jwtauth/v5"
)

// Authorizer is middleware to enforce Casbin authorization.
func Authorizer(enforcer *casbin.Enforcer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			_, claims, err := jwtauth.FromContext(req.Context())
			if err != nil {
				http.Error(res, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)

				return
			}

			user, ok := claims["sub"].(string)
			if !ok {
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
