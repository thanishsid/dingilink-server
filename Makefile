
run:
	go run github.com/joho/godotenv/cmd/godotenv@latest -f dev.env go run cmd/dingilink-server/main.go

gensql:
	sqlc generate

gengql:
	go run github.com/99designs/gqlgen

migrate-dry-run-local:
	atlas schema apply --url "postgres://thanish@localhost:5432/dingilink?sslmode=disable" --file "internal/db/schema/schema.sql" --dry-run --dev-url "postgres://thanish@localhost:5432/postgres?sslmode=disable"

migrate-local:
	atlas schema apply --url "postgres://thanish@localhost:5432/dingilink?sslmode=disable" --file "internal/db/schema/schema.sql" --dev-url "postgres://thanish@localhost:5432/postgres?sslmode=disable"