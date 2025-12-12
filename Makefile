COVERAGE_OUT ?= bin/coverage.out
COVERAGE_HTML ?= bin/coverage.html

DOCKER_COMPOSE_FILE := ./docker-compose.yml

bin:
	@mkdir -p bin

.PHONY: tidy
tidy:
	go mod tidy
	go fmt ./...

.PHONY: test
test: bin
	go test -v -race \
		-coverprofile=$(COVERAGE_OUT) \
		./...

.PHONY: cover
cover: test
	go tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@coverage=$$(go tool cover -func=$(COVERAGE_OUT) | grep total | awk '{print $$3}'); \
	echo "Coverage: $${coverage}"

.PHONY: worker.start
worker.start:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d
	go run cmd/worker/main.go

.PHONY: generate.mocks
generate.mocks:
	mockery --config ./.mockery.yml