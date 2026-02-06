# Northstar Group Demo - Makefile
# Simple build helper for democtl

.PHONY: build-democtl
build-democtl: ## Build democtl binary
	@echo "Building democtl..."
	@mkdir -p bin
	@cd democtl && go build -o ../bin/democtl ./cmd/democtl
	@chmod +x bin/democtl
	@echo "✓ Built democtl at bin/democtl"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Add to PATH (optional): fish_add_path \$$PWD/bin"
	@echo "  2. Run: democtl setup"
	@echo "  3. Run: democtl verify"
	@echo "  4. Run: democtl run <track>/<slug>"

.PHONY: help
help: ## Show this help
	@echo "Northstar Group Demo"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  %-20s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""
	@echo "After building, use 'democtl --help' for all demo commands"

.DEFAULT_GOAL := help
