.PHONY: install dev

install:
	cd frontend && bun install
	cd backend && go mod tidy

dev:
	bash -c 'cd backend && go run main.go' & bash -c 'cd frontend && bun run dev --port 3001'