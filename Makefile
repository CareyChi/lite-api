.PHONY: infra-up infra-down dev-api test vet web-dev web-check

infra-up:
	docker compose up -d

infra-down:
	docker compose down

dev-api:
	go run ./cmd/lite-api

test:
	go test ./...

vet:
	go vet ./...

web-dev:
	cd web && npm run dev

web-check:
	cd web && npm run typecheck && npm run build
