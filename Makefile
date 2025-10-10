.PHONY: install dev

install:
	cd frontend && bun install
	cd backend && go mod tidy

dev:
	bash -c 'cd backend && air' & bash -c 'cd frontend && bun run dev --port 3001'