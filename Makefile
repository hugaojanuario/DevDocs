COMPOSE = docker compose --env-file .env -f deployments/docker-compose.yml
.PHONY: build test lint up down migrate cli generate integration release
build:
	mkdir -p bin
	go build -trimpath -o bin/devdocs-api ./cmd/api
	go build -trimpath -o bin/devdocs ./cmd/devdocs
test:
	go test -race ./...
lint:
	go vet ./...
	@test -z "$$(gofmt -l cmd internal cli tests)"
up:
	$(COMPOSE) up -d --build --wait
down:
	$(COMPOSE) down
migrate:
	$(COMPOSE) run --rm devdocs-migrate
cli:
	go run ./cmd/devdocs $(ARGS)
generate:
	sqlc generate -f db/sqlc.yaml
integration:
	$(COMPOSE) --profile test run --build --rm devdocs-tests
release:
	sh scripts/build-cli.sh
