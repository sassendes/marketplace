VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
IMAGE ?= ghcr.io/sassendes/marketplace
TEST_DATABASE_URL ?= postgres://postgres:postgres@127.0.0.1:5432/marketplace_test?sslmode=disable

.PHONY: build web image test test-go test-rust lint run-db

web:
	cd web && npm ci --no-audit --no-fund && npm run build

build: web
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o bin/marketplace ./cmd/marketplace
	cd search-engine && cargo build --release --locked
	cd fraud-engine && cargo build --release --locked

image:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE):$(VERSION) -t $(IMAGE):latest .

test: test-go test-rust

test-go:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -race -count=1 ./...

test-rust:
	cd search-engine && cargo test --release --locked
	cd fraud-engine && cargo test --release --locked

lint:
	go vet ./...
	test -z "$$(gofmt -l $$(git ls-files '*.go'))"

run-db:
	docker run --rm -d --name marketplace-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=marketplace -p 5432:5432 postgres:16-alpine
	until docker exec marketplace-db pg_isready -U postgres >/dev/null 2>&1; do sleep 1; done
	docker exec marketplace-db psql -U postgres -c 'CREATE DATABASE marketplace_test'
