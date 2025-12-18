APP_NAME=bookture-server
UPLOAD_DIR=server/uploads

.PHONY: docker-run local-run docker-down clean restart restart-clean

docker-run:
	@echo "Starting Docker..."
	docker compose up --build

local-run:
	@echo "Starting Local..."
	$(MAKE) -C server run

docker-down:
	@echo "Stopping Docker..."
	docker compose down -v

clean: docker-down
	@echo "Cleaning uploads..."
	sudo rm -rf $(UPLOAD_DIR)

restart:
	@echo "Restarting with old data..."
	docker compose down
	docker compose up --build

restart-clean: clean
	@echo "Restarting with new data..."
	docker compose up --build
