# Testing Guide

This document describes the testing strategy and how to run tests for the JWT Application.

## Test Structure

The application has two types of tests:

### 1. Unit Tests
Unit tests test individual components in isolation without requiring external dependencies like databases or Redis.

**Location:** `*_test.go` files in each package
- `utils/utils_test.go` - Tests for utility functions (JWT generation, password hashing, etc.)

**Run unit tests:**
```bash
make test-unit
```

### 2. Integration Tests
Integration tests test the full application stack including HTTP endpoints, database interactions, and Redis caching.

**Location:** `integration_test.go`
- Full API endpoint testing
- Database operations
- Redis caching
- Complete user lifecycle tests

**Run integration tests:**
```bash
# With local services
make test-integration

# With Docker environment
make test-docker
```

## Running Tests

### Quick Test (Unit Tests Only)
```bash
make test-unit
```

### All Tests (Unit + Integration)
```bash
make test-all
```

### With Coverage Report
```bash
make test-coverage
# Opens coverage.html in your browser
```

### Using Docker Test Environment
The application includes a complete Docker test environment:

```bash
# Start test services (Postgres + Redis)
make test-docker-up

# Run tests
make test-docker

# Stop test services
make test-docker-down
```

### Complete Docker Test Run
```bash
make test-docker-run
```
This command:
1. Builds the test image
2. Starts Postgres and Redis
3. Runs all tests
4. Stops and cleans up

## Test Environment Configuration

### Local Testing
Copy `.env.test` to `.env` or export environment variables:

```bash
export APP_PORT=8080
export DB_HOST=localhost
export DB_USER=jwt-test-user
export DB_PASSWORD=jwt-test-password
export DB_PORT=5433
export DB_NAME=jwtapp-test
export DB_DIALECT=postgres
export REDIS_HOST=localhost
export REDIS_PORT=6380
export REDIS_PASSWORD=""
export SECRET=test-secret-key-for-jwt-signing
```

### Docker Testing
The `docker-compose.test.yml` file configures:
- PostgreSQL 15 on port 5433
- Redis 7 on port 6380
- Isolated test network
- Health checks for all services

## CI/CD Pipeline

### GitHub Actions
The `.github/workflows/ci.yml` workflow runs on every push and pull request:

1. **Unit Tests** - Fast tests without external dependencies
2. **Integration Tests** - Full API tests with Postgres and Redis services
3. **Build** - Compiles the application
4. **Coverage Report** - Generates code coverage metrics

### GoCD Pipeline
The `deploy-pipeline.gocd.yaml` defines:

**Stage 1: Test**
- Job: `run-unit-tests` - Executes unit tests with coverage
- Job: `run-integration-tests` - Executes integration tests with Docker

**Stage 2: Build**
- Job: `build-app` - Builds production binary

## Test Coverage

View coverage reports:

```bash
# Generate HTML coverage report
make test-coverage

# View in terminal
go test -cover ./...
```

## Writing Tests

### Unit Test Example
```go
func TestGenerateToken(t *testing.T) {
    os.Setenv("SECRET", "test-secret")
    defer os.Unsetenv("SECRET")

    user := models.User{ID: 1, Email: "test@example.com"}
    token, err := GenerateToken(user)

    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if token == "" {
        t.Error("Expected token, got empty string")
    }
}
```

### Integration Test Example
```go
func TestSignupAPI(t *testing.T) {
    clearTable()

    payload := `{"email":"test@example.com", "password":"password123"}`
    req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer([]byte(payload)))
    req.Header.Set("Content-Type", "application/json")

    response := executeRequest(req)
    checkResponseCode(t, http.StatusCreated, response.Code)
}
```

## Continuous Integration

### Pre-commit Checks
Before committing, run:
```bash
make test-all
```

### CI Commands
Commands used in CI/CD pipelines:
```bash
make ci-test              # Unit tests for CI
make ci-test-integration  # Integration tests with Docker for CI
```

## Troubleshooting

### Tests Fail - Database Connection
Ensure Postgres is running:
```bash
docker ps | grep postgres-test
# If not running:
make test-docker-up
```

### Tests Fail - Redis Connection
Ensure Redis is running:
```bash
docker ps | grep redis-test
# If not running:
make test-docker-up
```

### Clean Test Cache
```bash
make clean
```

### Port Conflicts
If ports 5433 or 6380 are in use:
```bash
# Stop existing containers
make test-docker-down

# Check what's using the ports
lsof -i :5433
lsof -i :6380
```

## Test Data Management

### Cleanup
Integration tests automatically:
- Clean the database before each test suite
- Clean the database after all tests complete
- Reset sequences to ensure consistent IDs

### Manual Cleanup
```bash
# Connect to test database
docker exec -it jwtapp-postgres-test psql -U jwt-test-user -d jwtapp-test

# Run cleanup
DELETE FROM users;
ALTER SEQUENCE users_id_seq RESTART WITH 1;
```

## Performance

- **Unit Tests**: < 1 second
- **Integration Tests**: ~5-10 seconds (includes service startup)
- **Full Test Suite**: ~10-15 seconds

## Best Practices

1. **Keep unit tests fast** - No external dependencies
2. **Isolate integration tests** - Use build tags
3. **Clean up after tests** - Reset database state
4. **Use table-driven tests** - Test multiple scenarios
5. **Mock external services** - In unit tests
6. **Test error cases** - Not just happy paths
7. **Maintain high coverage** - Aim for >80%
