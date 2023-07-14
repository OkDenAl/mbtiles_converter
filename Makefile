.PHONY: start-db stop-db
start-db:
	docker compose up -d --build

stop-db:
	docker compose down

cover:
	go test -v -coverpkg./... -coverprofile report.out -covermode=atomic./...