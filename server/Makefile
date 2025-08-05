.PHONY: build run clean tidy test deps run-race build-prod all build-kali-container check-kali-container setup

# Build the application
build:
	go build -o gogogajeto .

# Run the application
run:
	go run .

# Clean build artifacts
clean:
	rm -f gogogajeto

# Update dependencies
tidy:
	go mod tidy

# Run tests
test:
	go test ./...

# Install dependencies
deps:
	go mod download

# Run with race detection
run-race:
	go run -race .

# Build for production
build-prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gogogajeto .

# Build Kali Linux container with security tools
build-kali-container:
	@echo "🐳 Building Kali Linux security tools container..."
	cd docker && ./build-kali.sh

# Build Kali container without cache (clean build)
build-kali-container-clean:
	@echo "🐳 Building Kali Linux container (clean)..."
	cd docker && ./build-kali.sh --no-cache

# Build Kali container in background
build-kali-container-bg:
	@echo "🐳 Building Kali Linux container in background..."
	cd docker && nohup ./build-kali.sh > build.log 2>&1 &
	@echo "📄 Monitor progress with: tail -f docker/build.log"

# Check if Kali container exists and is working
check-kali-container:
	@echo "🔍 Checking Kali container..."
	@if docker images -q gogogadgeto/kali-tools:latest >/dev/null 2>&1; then \
		echo "✅ Container image exists"; \
		if docker run --rm gogogadgeto/kali-tools:latest nmap --version >/dev/null 2>&1; then \
			echo "✅ Container is working (nmap test passed)"; \
		else \
			echo "❌ Container exists but nmap test failed"; \
			echo "💡 Try: make build-kali-container-clean"; \
		fi \
	else \
		echo "❌ Container image not found"; \
		echo "💡 Run: make build-kali-container"; \
	fi

# Setup everything (build container + app)
setup: build-kali-container build
	@echo "🎉 Setup complete!"

# Default target
all: tidy build 