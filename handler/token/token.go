// Package token contains token generation implementations.
package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v4"
)

// Errors.
var (
	ErrEmptySecret = errors.New("empty secret")
)

// JWT is used to generate jwt tokens.
type JWT struct {
	jwtauth         *jwtauth.JWTAuth
	issuer          string
	authTokenExp    time.Duration
	refreshTokenExp time.Duration
}

// NewJWT creates new jwt.
func NewJWT(issuer, secret string, authTokenExp, refreshTokenExp time.Duration) (*JWT, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	return &JWT{
		jwtauth:         jwtauth.New("HS256", []byte(secret), nil),
		issuer:          issuer,
		authTokenExp:    authTokenExp,
		refreshTokenExp: refreshTokenExp,
	}, nil
}

// Tokens generates auth and refresh jwt tokens.
func (j *JWT) Tokens(userID, requestID string) (authToken, refreshToken string, err error) { //nolint:nonamedreturns
	now := time.Now()

	if _, authToken, err = j.jwtauth.Encode(jwt.MapClaims{
		"iss": j.issuer,
		"exp": now.Add(j.authTokenExp).Unix(),
		"iat": now.Unix(),
		"jti": requestID,
		"sub": userID,
	}); err != nil {
		return "", "", fmt.Errorf("encode auth token: %w", err)
	}

	if _, refreshToken, err = j.jwtauth.Encode(jwt.MapClaims{
		"iss": j.issuer,
		"exp": now.Add(j.refreshTokenExp).Unix(),
		"iat": now.Unix(),
		"nbf": now.Add(j.authTokenExp).Unix(),
		"jti": requestID,
		"sub": userID,
	}); err != nil {
		return "", "", fmt.Errorf("encode refresh token: %w", err)
	}

	return authToken, refreshToken, nil
}

// JWTAuth returns jwt auth.
func (j *JWT) JWTAuth() *jwtauth.JWTAuth {
	return j.jwtauth
}
