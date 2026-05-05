.PHONY: build run api worker docker-up docker-down migrate test

build:
	go build -o bin/api ./cmd/api
	go build -o bin/worker ./cmd/worker

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	psql postgres://vturb:vturb@localhost:5433/vturb -f migrations/001_create_videos.sql

test:
	go test ./...

dev:
	make docker-up
	make migrate
	make run-api

.DEFAULT_GOAL := build
