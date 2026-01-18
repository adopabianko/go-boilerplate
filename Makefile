
# Variables
APP_NAME=go-boilerplate
DB_URL=postgres://postgres:password@localhost:5432/boilerplate_db?sslmode=disable

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


