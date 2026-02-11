# Load .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Variables
APP_NAME=go-boilerplate
DB_URL=postgres://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSLMODE)

.PHONY: run build test clean docker-up docker-down migrate-create migrate-up migrate-down

run:
	go run cmd/api/main.go

run-dummy-grpc:
	go run cmd/dummy_grpc_server/main.go

air:
	air

build:
	go build -o tmp/$(APP_NAME) cmd/api/main.go

test:
	go test -v ./...

clean:
	rm -rf tmp

docker-up:
	docker compose up -d

docker-down:
	docker compose down

swagger:
	swag init -g cmd/api/main.go

# Migration Commands (requires golang-migrate CLI)
# Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

migrate-up:
	migrate -path migrations -database "$(DB_URL)" -verbose up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" -verbose down

migrate-force:
	@read -p "Enter version to force: " version; \
	migrate -path migrations -database "$(DB_URL)" -verbose force $$version

proto:
	protoc --go_out=. --go_opt=module=go-boilerplate \
	--go-grpc_out=. --go-grpc_opt=module=go-boilerplate \
	api/proto/payment/payment.proto



cert:
	mkdir -p certs
	openssl genrsa -out certs/private.pem 2048
	openssl rsa -in certs/private.pem -pubout -out certs/public.pem
