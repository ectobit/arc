# arc

[![Build Status](https://github.com/ectobit/arc/workflows/build/badge.svg)](https://github.com/ectobit/arc/actions)
[![Go Reference](https://pkg.go.dev/badge/go.ectobit.com/arc.svg)](https://pkg.go.dev/go.ectobit.com/arc)
[![Go Report](https://goreportcard.com/badge/go.ectobit.com/arc)](https://goreportcard.com/report/go.ectobit.com/arc)
[![License](https://img.shields.io/badge/license-BSD--2--Clause--Patent-orange.svg)](https://github.com/ectobit/arc/blob/main/LICENSE)

REST API in Go user accounting and authentication.

## Features

- [x] User registration
- [x] Send mail for account activation
- [x] Account activation
- [ ] Send mail for password reset
- [ ] Password reset
- [x] User login
- [x] Password strength check
- [x] JWT based authentication
- [ ] Good test coverage (just 9% at the moment)
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
- `make test-cov` displays test coverage

## [Swagger specification](http://localhost:3000/)

## Links related to future tasks

- [GitHub Actions - Running jobs in containers](https://docs.github.com/en/actions/using-containerized-services/creating-postgresql-service-containers)
- [Control startup and shutdown order in Compose](https://docs.docker.com/compose/startup-order/)
