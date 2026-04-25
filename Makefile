BINARY=meeting-room-booking
CLI=booking-cli
DB_URL?=postgres://postgres:postgres@localhost:5432/meeting_room?sslmode=disable

export DATABASE_URL=$(DB_URL)

.PHONY: run cli build install migrate-up migrate-down fmt test clean help

run:
	go run ./cmd/launcher

cli:
	go run ./cmd/booking-cli

build:
	go build -o $(BINARY) ./cmd/launcher
	go build -o $(CLI) ./cmd/booking-cli

install:
	go mod tidy

migrate-up:
	psql "$(DB_URL)" -f db/migrations/1_meeting_room_up.sql

migrate-down:
	psql "$(DB_URL)" -f db/migrations/1_meeting_room_down.sql

fmt:
	go fmt ./...

test:
	go test ./...

clean:
	rm -f $(BINARY) $(CLI)

help:
	@echo "make run            start the API server"
	@echo "make cli            run the CLI without arguments"
	@echo "make build          build the API and CLI binaries"
	@echo "make install        tidy module dependencies"
	@echo "make migrate-up     apply database schema"
	@echo "make migrate-down   drop database schema"
	@echo "make fmt            format Go files"
	@echo "make test           run Go tests"
	@echo "make clean          remove built binaries"
