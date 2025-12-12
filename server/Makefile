APP_NAME=bookture-server
BIN_DIR=bin
CMD_DIR=./cmd/main.go

.PHONY: all build run clean seed

all: build

build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)
	@echo "Build complete. Binary located at $(BIN_DIR)/$(APP_NAME)"

run: build
	@echo "Starting $(APP_NAME)..."
	./$(BIN_DIR)/$(APP_NAME)

clean:
	@echo "Cleaning up..."
	rm -rf $(BIN_DIR)
	@echo "Clean complete."

migrate:
	@echo "Migrating tables..."
	go run ./migration/main.go

seed:
	@echo "Seeding tables..."
	go run ./seed/main.go