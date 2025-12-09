.PHONY: help docker-up docker-down migrate build-api build-processor run-api run-processor clean setup

help:
	@echo "Order Management Platform - Commands"
	@echo ""
	@echo "Setup & Infrastructure:"
	@echo "  make docker-up          Start PostgreSQL, Kafka, Prometheus"
	@echo "  make docker-down        Stop all containers"
	@echo "  make migrate            Run database migrations"
	@echo ""
	@echo "Build:"
	@echo "  make build-api          Build order-api binary"
	@echo "  make build-processor    Build order-processor binary"
	@echo "  make build-all          Build all binaries"
	@echo ""
	@echo "Run:"
	@echo "  make run-api            Run order-api service"
	@echo "  make run-processor      Run order-processor service"
	@echo ""
	@echo "Full Setup:"
	@echo "  make setup              Complete setup (docker + migrate)"
	@echo "  make clean              Clean build artifacts"

docker-up:
	docker-compose up -d
	@echo "⏳ Waiting for services..."
	@sleep 10
	@echo "✓ Services started"

docker-down:
	docker-compose down -v
	@echo "✓ Services stopped"

migrate:
	go run cmd/migrations/main.go

build-api:
	CGO_ENABLED=1 go build -o bin/order-api ./cmd/order-api

build-processor:
	CGO_ENABLED=1 go build -o bin/order-processor ./cmd/order-processor

build-all: build-api build-processor

run-api: build-api
	./bin/order-api

run-processor: build-processor
	./bin/order-processor

clean:
	rm -rf bin/

setup: docker-up migrate
	@echo ""
	@echo "✓ Setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  Terminal 1: make run-api"
	@echo "  Terminal 2: make run-processor"
	@echo ""
	@echo "Then test with:"
	@echo "  curl -X POST http://localhost:8080/api/v1/orders ..."
