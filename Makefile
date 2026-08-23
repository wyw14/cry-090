APP=salsa-circle
GO=go
run:
	$(GO) run ./cmd/server
build:
	$(GO) build ./...
test:
	$(GO) test ./...
race:
	$(GO) test -race ./...
vet:
	$(GO) vet ./...
fmt:
	gofmt -w .
frontend-install:
	npm ci
frontend-test:
	npm run test
frontend-build:
	npm run build
docker:
	docker compose up --build
migrate:
	@echo "Apply migrations/001_init.sql and migrations/002_seed.sql with your PostgreSQL client"
