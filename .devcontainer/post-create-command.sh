#!/bin/bash
set -e

echo "🚀 Azure-Nuke DevContainer Initialization"
echo "=========================================="

# Cache Go dependencies
echo "📦 Caching Go dependencies..."
go mod download

# Initialize git hooks with pre-commit
echo "🔧 Setting up git hooks with pre-commit..."
if command -v pre-commit &> /dev/null; then
    pre-commit install || echo "⚠️  pre-commit install had issues, but continuing..."
    echo "✅ Git hooks installed"
else
    echo "⚠️  pre-commit not found in PATH (feature may have failed), skipping hook setup"
fi

# Clean golangci-lint cache
echo "🧹 Cleaning golangci-lint cache..."
golangci-lint cache clean || echo "⚠️  golangci-lint not yet available"

# Validate container runtime socket and provide guidance for Docker/Colima/Podman.
bash .devcontainer/container-runtime-preflight.sh

echo ""
echo "✨ DevContainer initialization complete!"
echo ""
echo "📚 Quick Start Guide:"
echo "  - Run tests:              go test ./..."
echo "  - Run linting:            golangci-lint run"
echo "  - Build snapshot:         make build"
echo "  - Build docs:             make docs"
echo "  - Serve docs (port 8080): make docs-serve"
echo ""
echo "🔧 Development Tips:"
echo "  - Docker socket is mounted; you can build container images"
echo "  - pre-commit is installed via Dev Container feature"
echo "  - Delve debugger available for step-through debugging (F5)"
echo "  - For Azure SDK auth: set AZURE_* env vars or mount ~/.azure"
echo ""
