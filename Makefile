.PHONY: gen-swagger lint start stop test test-all test-cov

gen-swagger:
	@swag init

lint:
	@golangci-lint run --exclude-use-default=false --enable-all \
		--disable golint \
		--disable interfacer \
		--disable scopelint \
		--disable maligned

start:
	@docker-compose up --build

stop:
	@docker-compose down

test:
	@go test -race -short ./...

test-all:
	PGPASSWORD=arc psql -U postgres -h localhost -d test -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	migrate -path=migrations -database='postgres://postgres:arc@localhost/test?sslmode=disable&query' up
	ARC_DB_HOST=localhost go test -race ./...

test-cov:
	PGPASSWORD=arc psql -U postgres -h localhost -d test -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp"'
	ARC_DB_HOST=localhost go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
