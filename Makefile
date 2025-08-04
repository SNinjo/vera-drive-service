.PHONY: wire start dev build prod test test-cov test-cov-html access-db insert-user-to-db migrate-create migrate-up migrate-down migrate-force

SHELL := /bin/bash
ENV := source .env &&

wire:
	go run github.com/google/wire/cmd/wire ./internal/app/

start:
	go run ./cmd/...

dev:
	reflex -s -r '(\.go$$|^\.env$$)' -R '(_gen\.go$$)' -- sh -c 'make wire && make start'

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -a -installsuffix cgo -o bin/main ./cmd/...
	chmod +x bin/main

prod:
	$(ENV) GIN_MODE=release bin/main

path ?= ./...
test:
	@cmd="go test $(path)"; \
	if echo "$(MAKECMDGOALS)" | grep -qw verbose; then \
		cmd="$$cmd -v"; \
	fi; \
	if [ -n "$(strip $(name))" ]; then \
		cmd="$$cmd -run $(name)"; \
	fi; \
	echo $$cmd; \
	$$cmd

test-cov:
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out

test-cov-html:
	go tool cover -html=cover.out

access-db:
	$(ENV) psql "$$DB_URL"

insert-user-to-db:
	$(ENV) psql "$$DB_URL" -c "INSERT INTO users (email) VALUES ('$(email)');"

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-up:
	$(ENV) migrate -path migrations -database "$$DB_URL" up

migrate-down:
	$(ENV) migrate -path migrations -database "$$DB_URL" down 1

migrate-force:
	$(ENV) migrate -path migrations -database "$$DB_URL" force $(version)
