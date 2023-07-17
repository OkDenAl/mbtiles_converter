.PHONY: start-db stop-db
start:
	docker compose up -d --build

stop:
	docker compose down

cover:
	go test -v -coverpkg./... -coverprofile report.out -covermode=atomic./...