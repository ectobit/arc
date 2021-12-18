.PHONY: gen-swagger lint start stop test test-cov

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
	@go test -race ./...

test-cov:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out
