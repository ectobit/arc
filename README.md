# arc

[![Build Status](https://github.com/ectobit/arc/workflows/build/badge.svg)](https://github.com/ectobit/arc/actions)
[![Go Reference](https://pkg.go.dev/badge/go.ectobit.com/arc.svg)](https://pkg.go.dev/go.ectobit.com/arc)
[![Go Report](https://goreportcard.com/badge/go.ectobit.com/arc)](https://goreportcard.com/report/go.ectobit.com/arc)
![Test Coverage](https://img.shields.io/badge/coverage-47.6%25-brightgreen?style=flat&logo=go)
[![License](https://img.shields.io/badge/license-BSD--2--Clause--Patent-orange.svg)](https://github.com/ectobit/arc/blob/main/LICENSE)

REST API in Go user accounting and authentication.

## Features

- [x] User registration, send activation link per email, user account activation
- [x] Password strength check
- [x] Request password reset, send mail with password reset token, password reset
- [x] User login
- [ ] Refresh token
- [x] JWT based authentication
- [x] Tested
- [x] Swagger specification
- [ ] Authorization
- [ ] Send messages to message queue
- [ ] SSE from message queue

## Contribution

- `make gen-swagger` regenerates swagger specification
- `make lint` lints the project
- `make start` starts docker-compose stack
- `make stop` stops docker-compose stack
- `make test` runs unit tests
- `make test-all` runs integration tests (requires docker-stack to be up)
- `make test-cov` displays test coverage (requires docker-stack to be up)

## [Swagger specification](http://localhost:3000/)

## Tips

If token should be parsed from query as well:

```
r.Use(func(next http.Handler) http.Handler {
    return jwtauth.Verify(s.jwtAuth, jwtauth.TokenFromQuery, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie)(next)
})
```
