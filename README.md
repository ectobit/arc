# arc

[![Build Status](https://github.com/ectobit/arc/workflows/build/badge.svg)](https://github.com/ectobit/arc/actions)
[![Go Reference](https://pkg.go.dev/badge/go.ectobit.com/arc.svg)](https://pkg.go.dev/go.ectobit.com/arc)
[![Go Report](https://goreportcard.com/badge/go.ectobit.com/arc)](https://goreportcard.com/report/go.ectobit.com/arc)
[![License](https://img.shields.io/badge/license-BSD--2--Clause--Patent-orange.svg)](https://github.com/ectobit/arc/blob/main/LICENSE)

REST API in Go providing user registration, account activation, login, password reset and JWT based authentication.

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
- [ ] Swagger specification
- [ ] Authorization
- [ ] Send messages to message queue
- [ ] SSE from message queue

## Links related to future tasks

- [GitHub Actions - Running jobs in containers](https://docs.github.com/en/actions/using-containerized-services/creating-postgresql-service-containers)
- [Control startup and shutdown order in Compose](https://docs.docker.com/compose/startup-order/)
