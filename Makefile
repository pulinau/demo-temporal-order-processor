.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: worker.start
worker.start:
	go run cmd/worker/main.go