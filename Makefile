.PHONY: help setup build-kali build-ui build-server build-all start-server start-ui start-all stop-all clean-all dev prod test check-kali check-deps install-deps

# Default target
help:
	@echo "🔍 GoGoGadgeto Security Tool - Build System"
	@echo ""
	@echo "🚀 Quick Start Commands:"
	@echo "  make setup        - Complete project setup (containers + dependencies)"
	@echo "  make dev          - Start development environment (both servers)"
	@echo "  make prod         - Build and start production environment"
	@echo "  make stop-all     - Stop all running services"
	@echo ""
	@echo "🐳 Container Management:"
	@echo "  make build-kali   - Build Kali Linux security tools container"
	@echo "  make check-kali   - Check if Kali container exists and works"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  make build-ui     - Build React frontend"
	@echo "  make build-server - Build Go backend"
	@echo "  make build-all    - Build everything"
	@echo ""
	@echo "🖥️  Server Management:"
	@echo "  make start-server - Start Go backend server"
	@echo "  make start-ui     - Start React development server"
	@echo "  make start-all    - Start both servers"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  make clean-all    - Clean all build artifacts"
	@echo "  make check-deps   - Check all dependencies"
	@echo "  make install-deps - Install all dependencies"

# Project setup - everything needed to get started
setup: check-deps install-deps build-kali
	@echo "🎉 GoGoGadgeto setup complete!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Add your OpenAI API key to server/.env"
	@echo "  2. Run 'make dev' to start development servers"

# Development environment - start both servers
dev: build-ui
	@echo "🚀 Starting development environment..."
	@echo "📝 Make sure you have your OpenAI API key in server/.env"
	@echo ""
	@echo "Starting Go server (http://localhost:8080)..."
	@cd server && make run &
	@sleep 3
	@echo "Starting React UI (http://localhost:5173)..."
	@cd ui && npm start &
	@echo ""
	@echo "✅ Development servers started!"
	@echo "📱 UI: http://localhost:5173"
	@echo "🔧 API: http://localhost:8080"
	@echo ""
	@echo "To stop: make stop-all"

# Production environment
prod: build-all
	@echo "🏭 Starting production environment..."
	@cd server && ./gogogajeto &
	@echo "✅ Production server started on http://localhost:8080"

# Container build commands
build-kali:
	@echo "🐳 Building Kali Linux security tools container..."
	@cd server && make build-kali-container

check-kali:
	@echo "🔍 Checking Kali container status..."
	@cd server && make check-kali-container

# Build commands
build-ui:
	@echo "🎨 Building React frontend..."
	@cd ui && npm run build

build-server:
	@echo "⚙️ Building Go backend..."
	@cd server && make build

build-all: build-ui build-server
	@echo "✅ All components built successfully!"

# Server management
start-server:
	@echo "🚀 Starting Go backend server..."
	@cd server && make run &

start-ui:
	@echo "🎨 Starting React development server..."
	@cd ui && npm start &

start-all: start-server start-ui
	@echo "✅ Both servers started!"

# Stop all services
stop-all:
	@echo "🛑 Stopping all services..."
	@pkill -f "go run" || true
	@pkill -f "npm start" || true
	@pkill -f "gogogajeto" || true
	@pkill -f "vite" || true
	@echo "✅ All services stopped"

# Clean up
clean-all:
	@echo "🧹 Cleaning all build artifacts..."
	@cd server && make clean
	@cd ui && rm -rf dist/ node_modules/.cache/
	@echo "✅ Clean complete"

# Dependency management
check-deps:
	@echo "🔍 Checking dependencies..."
	@echo "Checking Go..."
	@go version || (echo "❌ Go not found. Please install Go 1.19+" && exit 1)
	@echo "Checking Node.js..."
	@node --version || (echo "❌ Node.js not found. Please install Node.js 16+" && exit 1)
	@echo "Checking Docker..."
	@docker --version || (echo "❌ Docker not found. Please install Docker" && exit 1)
	@echo "Checking Make..."
	@make --version || (echo "❌ Make not found. Please install Make" && exit 1)
	@echo "✅ All dependencies found!"

install-deps:
	@echo "📦 Installing dependencies..."
	@echo "Installing Go dependencies..."
	@cd server && go mod download
	@echo "Installing Node.js dependencies..."
	@cd ui && npm install
	@echo "✅ Dependencies installed!"

# Test commands
test:
	@echo "🧪 Running tests..."
	@cd server && make test
	@cd ui && npm test --passWithNoTests
	@echo "✅ All tests passed!"

# Quick status check
status:
	@echo "📊 GoGoGadgeto Status:"
	@echo ""
	@echo "🐳 Kali Container:"
	@cd server && make check-kali-container 2>/dev/null || echo "❌ Not available"
	@echo ""
	@echo "🔄 Running Processes:"
	@pgrep -f "go run" >/dev/null && echo "✅ Go server running" || echo "❌ Go server not running"
	@pgrep -f "vite" >/dev/null && echo "✅ UI server running" || echo "❌ UI server not running" 