.PHONY: gen-swagger lint start stop test test-cov

gen-swagger:
	@swag init

lint:
	@golangci-lint run

start:
	@docker-compose up --build

stop:
	@docker-compose down

test:
	@go test -race ./...

test-cov:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
