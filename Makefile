# Jaqen NewGen Tool - Simple Build System

APP_NAME = jaqen-newgen-tool

.PHONY: build dev clean help

# Build the application
build:
	go build -o $(APP_NAME) .

# Build and run for development
dev:
	make build && ./$(APP_NAME)

# Clean build artifacts
clean:
	rm -f $(APP_NAME)

# Show help
help:
	@echo "Jaqen NewGen Tool - Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build    - Build the application"
	@echo "  dev      - Build and run for development"
	@echo "  clean    - Clean build artifacts"
	@echo "  help     - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build    # Build the application"
	@echo "  make dev      # Build and run locally"