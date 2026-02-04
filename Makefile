.PHONY: dev build run clean

dev: ## Development - run locally without Docker
	go run main.go


build: ## Build the Docker image
	docker build --platform linux/arm64 -t ip-clock-app .

run: ## Run the Docker container
	docker run -p 8080:8080 ip-clock-app

up: build run ## Build and run the Docker container

stop: ## Stop the Docker container
	docker stop $$(docker ps -q --filter ancestor=ip-clock-app) 2>/dev/null || true

clean: ## Remove the Docker image
	docker rmi ip-clock-app 2>/dev/null || true

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
