.PHONY: start stop cover

start:
	docker compose up -d --build

stop:
	docker compose down

cover:
	go test -v -coverpkg=./... -coverprofile report.out -covermode=atomic ./...
	grep -v -E -- '*mocks|vector_tile|config|cmd|repository|logging'  report.out > report1.out
	go tool cover -func=report1.out