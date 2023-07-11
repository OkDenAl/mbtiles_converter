.PHONY: start-db stop-db
start-db:
	docker compose up -d --build

stop-db:
	docker compose down