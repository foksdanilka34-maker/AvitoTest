.PHONY:  run format-all lint-run

run:
	gofmt -s -w .
	go run ./cmd/main.go

lint-run:
	golangci-lint run ./...