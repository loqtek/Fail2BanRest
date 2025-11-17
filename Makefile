.PHONY: build run clean test install

build:
	@echo "Building fail2restV2..."
	@go build -o fail2restV2 ./cmd/server
	@go build -o hash-password ./cmd/hash-password
	@echo "Build complete!"

run: build
	@echo "Running fail2restV2..."
	@./fail2restV2

clean:
	@echo "Cleaning..."
	@rm -f fail2restV2
	@go clean
	@echo "Clean complete!"

test:
	@echo "Running tests..."
	@go test ./...

install: build
	@echo "Installing fail2restV2..."
	@sudo cp fail2restV2 /usr/local/bin/
	@echo "Installation complete!"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated!"

