dev:
	docker compose -f docker-compose.yaml -f docker-compose.dev.yaml up --build

up:
	docker compose up --build
