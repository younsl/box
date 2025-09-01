.PHONY: run start stop clean logs help volume-create volume-remove

CONTAINER_NAME := flatnotes
IMAGE := dullage/flatnotes:latest
PORT := 8080
VOLUME_NAME := flatnotes-data

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

volume-create: ## Create Docker volume for flatnotes data
	podman volume create $(VOLUME_NAME)

volume-remove: ## Remove Docker volume (WARNING: This will delete all data)
	podman volume rm $(VOLUME_NAME) || true

run: volume-create ## Run flatnotes container and open in browser
	@read -p "Enter username: " username; \
	read -s -p "Enter password: " password; \
	echo; \
	podman run -d \
		--name $(CONTAINER_NAME) \
		--restart unless-stopped \
		-e "PUID=1000" \
		-e "PGID=1000" \
		-e "FLATNOTES_AUTH_TYPE=password" \
		-e "FLATNOTES_USERNAME=$$username" \
		-e "FLATNOTES_PASSWORD=$$password" \
		-e "FLATNOTES_SECRET_KEY=aLongRandomSeriesOfCharacters" \
		-v "$(VOLUME_NAME):/data" \
		-p "$(PORT):8080" \
		$(IMAGE)
	@echo "Waiting for container to start..."
	@sleep 2 
	@open http://localhost:$(PORT)

start: ## Start existing container and open in browser
	podman start $(CONTAINER_NAME)
	@echo "Waiting for container to start..."
	@sleep 2
	@open http://localhost:$(PORT)

stop: ## Stop flatnotes container
	podman stop $(CONTAINER_NAME) || true

clean: stop ## Stop container, remove it and remove Docker volume
	podman rm $(CONTAINER_NAME) || true
	$(MAKE) volume-remove

logs: ## Show container logs
	podman logs -f $(CONTAINER_NAME)

restart: stop start ## Restart the container
