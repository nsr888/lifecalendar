.PHONY: run lint fmt

run:
	go run ./cmd

home:
	go run ./cmd ./config_home.toml

lint: ## golangci-lint v1.64.8 run from docker container
	@echo "Running golangci-lint..."
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v2.6.2 golangci-lint run --config .golangci.yaml

fmt:
	@echo "Formatting code..."
	gofmt -w -s .
