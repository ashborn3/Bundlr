run:
	go run ./cmd/server
build:
	go build -o bundlr_server ./cmd/server

# Too bad you can't ping localhost:5432 LMAO
reset-db:
	migrate -path migrations -database postgres://myuser:mypassword@localhost:5432/bundlr?sslmode=disable down -all
	migrate -path migrations -database postgres://myuser:mypassword@localhost:5432/bundlr?sslmode=disable up