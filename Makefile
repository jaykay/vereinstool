.PHONY: dev dev-backend dev-frontend sqlc migrate-up migrate-down build clean

# Start backend (air) + frontend (vite) in parallel
dev:
	@echo "Starting backend (air) and frontend (vite)..."
	@trap 'kill 0' EXIT; \
		$(MAKE) dev-backend & \
		$(MAKE) dev-frontend & \
		wait

dev-backend:
	cd backend && air

dev-frontend:
	cd frontend && npm run dev

# Database
sqlc:
	cd backend/db && sqlc generate

migrate-up:
	cd backend && migrate -path db/migrations -database "sqlite3://./dev.db" up

migrate-down:
	cd backend && migrate -path db/migrations -database "sqlite3://./dev.db" down 1

# Build
build:
	cd frontend && npm run build
	cd backend && CGO_ENABLED=1 go build -o vereinstool ./main.go

clean:
	rm -f backend/vereinstool backend/dev.db
	rm -rf backend/tmp frontend/build frontend/.svelte-kit
