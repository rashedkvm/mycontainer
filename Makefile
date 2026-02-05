ARCH := $(shell uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')
PLATFORM := linux/$(ARCH)
REGISTRY ?=
CONTAINER_ID ?= mycontainer-001

.PHONY: dev build run clean check-registry k8s-deploy k8s-delete

check-registry: ## Verify REGISTRY is set
	@if [ -z "$(REGISTRY)" ]; then \
		echo "Error: REGISTRY value not set"; \
		exit 1; \
	fi

dev: ## Development - run locally without Docker
	go run main.go

build: check-registry ## Build the Docker image
	docker build --platform $(PLATFORM) -t $(REGISTRY)/ip-clock-app .

push: check-registry build ## Build and Push the Docker image to the registry
	docker push $(REGISTRY)/ip-clock-app

run: check-registry ## Run the Docker container
	docker run -p 8080:8080 -e CONTAINER_ID=$(CONTAINER_ID) $(REGISTRY)/ip-clock-app

up: build run ## Build and run the Docker container

k8s-deploy: check-registry ## Deploy application to Kubernetes
	ytt -f config/k8s/ -v image.registry=$(REGISTRY) -v containerID=$(CONTAINER_ID) | kbld -f - | kapp -y deploy -a ip-clock-app -f -

k8s-delete: ## Delete the Kubernetes application
	kapp delete -a ip-clock-app

stop: ## Stop the Docker container
	docker stop $$(docker ps -q --filter ancestor=ip-clock-app) 2>/dev/null || true

clean: ## Remove the Docker image
	docker rmi ip-clock-app 2>/dev/null || true

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
