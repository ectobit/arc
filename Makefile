.PHONY: lint test test-all test-all-ci test-coverage start stop

lint:
	@golangci-lint run --exclude-use-default=false --enable-all \
		--disable golint \
		--disable interfacer \
		--disable scopelint \
		--disable maligned

test:
	@go test -short ./...

test-all:
	PGPASSWORD=arc psql -U postgres -h localhost -d test -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	ARC_DB_HOST=localhost go test ./...

test-all-ci:
	PGPASSWORD=arc psql -U postgres -h postgres -d test -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	migrate -path=migrations -database='postgres://postgres:arc@postgres/test?sslmode=disable&query' up
	ARC_DB_HOST=postgres go test ./...

test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out

start:
	@docker-compose up --build

stop:
	@docker-compose down
