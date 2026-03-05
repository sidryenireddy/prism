.PHONY: api frontend dev test lint migrate generate clean

api:
	cd api && go run ./cmd/server

frontend:
	cd frontend && npm run dev

dev:
	$(MAKE) -j2 api frontend

test:
	cd api && go test ./...
	cd frontend && npm test

lint:
	cd api && go vet ./...
	cd frontend && npm run lint

migrate:
	cd api && go run ./cmd/server -migrate

generate:
	oapi-codegen -generate types -o api/internal/models/openapi.go -package models spec/openapi.yaml

clean:
	cd api && go clean
	cd frontend && rm -rf .next node_modules

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down
