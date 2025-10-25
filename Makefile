# Database commands
postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=jwt-user -e POSTGRES_PASSWORD=123456 -d postgres:12-alpine

redis:
	docker run --name redis -p 6379:6379 -d redis:latest

createdb:
	docker exec -it postgres12 createdb --username=jwt-user --owner=jwt-user jwtdb

dropdb:
	docker exec -it postgres12 dropdb jwtdb

migrateup:
	migrate -path database/migration -database "postgresql://jwt-user:123456@localhost:5432/jwtdb?sslmode=disable" -verbose up

migratedown:
	migrate -path database/migration -database "postgresql://jwt-user:123456@localhost:5432/jwtdb?sslmode=disable" -verbose down

# Application commands
server:
	go run main.go

build:
	go build -o app .

# Test commands
test:
	go test -v -cover ./...

test-unit:
	@echo "Running unit tests..."
	go test -v -cover -short ./...

test-integration:
	@echo "Running integration tests..."
	go test -v -cover -tags=integration ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

test-all: test-unit test-integration

# Docker test environment commands
test-docker-up:
	@echo "Starting test environment..."
	docker-compose -f docker-compose.test.yml up -d postgres-test redis-test
	@echo "Waiting for services to be ready..."
	sleep 5

test-docker-down:
	@echo "Stopping test environment..."
	docker-compose -f docker-compose.test.yml down -v

test-docker-run:
	@echo "Running tests in Docker..."
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit test-runner

test-docker: test-docker-up
	@echo "Running tests with Docker environment..."
	sleep 3
	export $$(cat .env.test | xargs) && go test -v -tags=integration ./...
	$(MAKE) test-docker-down

# CI/CD test commands
ci-test:
	@echo "Running CI unit tests with coverage..."
	go test -v -coverprofile=coverage.out -covermode=atomic -short ./...

ci-test-integration: test-docker-up
	@echo "Running CI integration tests..."
	sleep 3
	export $$(cat .env.test | xargs) && go test -v -tags=integration -coverprofile=coverage-integration.out ./...
	$(MAKE) test-docker-down

# Clean commands
clean:
	rm -f app coverage.out coverage.html coverage-integration.out
	go clean -testcache

.PHONY: postgres redis createdb dropdb migrateup migratedown server build \
        test test-unit test-integration test-coverage test-all \
        test-docker-up test-docker-down test-docker-run test-docker \
        ci-test ci-test-integration clean