SHELL := /bin/bash
APP ?= quran-api
PORT ?= 8080
DB_PATH ?= quran.db

.PHONY: help
help:
	@echo "Targets: deps seed api web tui cli lint test sec vuln fmt docker.up docker.down precommit"

deps:
	go mod tidy

seed:
	go run ./scripts/seed.go

api:
	dotenvx run -- go run ./cmd/quran-api

web:
	dotenvx run -- go run ./cmd/quran-web

tui:
	go run ./cmd/quran-tui

cli:
	go run ./cmd/quran-cli

lint:
	staticcheck ./...

fmt:
	go fmt ./...

test:
	go test ./...

sec:
	gosec ./...

vuln:
	govulncheck ./...

docker.up:
	docker compose up -d --build

docker.down:
	docker compose down -v

precommit:
	pre-commit install
