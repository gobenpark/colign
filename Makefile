.PHONY: setup proto build run test lint clean dev

# Initial setup
setup:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
	cd web && npm install
	cd hocuspocus && npm install
	$(MAKE) proto

# Proto generation
proto:
	cd proto && buf generate

proto-lint:
	cd proto && buf lint

# Go
build:
	go build -o bin/api ./cmd/api
	go build -o bin/mcp ./cmd/mcp

run:
	go run ./cmd/api

test:
	go test ./...

lint:
	golangci-lint run ./...

# Frontend
web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

# Docker
up:
	docker compose up -d postgres redis

down:
	docker compose down

# All-in-one dev
dev: up run

clean:
	rm -rf bin/ gen/ web/src/gen/
