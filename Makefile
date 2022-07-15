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

test:
	go test -v -cover ./...

server:
	go run main.go


.PHONY: postgres redis createdb dropdb migrateup migratedown test server