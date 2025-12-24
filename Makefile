.PHONY: run lint fmt

run:
	go run ./cmd

home:
	go run ./cmd ./config_home.toml

lint:
	@echo "Running golangci-lint..."
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.6.2 golangci-lint run --config .golangci.yaml

.DEFAULT_GOAL := home
