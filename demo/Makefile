# Demo Makefile
# Main entry point for all demo commands
#
# Usage: make <target> [SCENARIO=<path>] [KUBE_CONTEXT=<context>]

.DEFAULT_GOAL := help

# Variables
SCENARIO ?=
STAGE ?=
TYPE ?=
KUBE_CONTEXT ?= kind-fider-demo
SHELL := /bin/bash

# Paths
DEMO_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
SHARED_DIR := $(DEMO_DIR)/shared
SRE_DIR := $(DEMO_DIR)/sre
ENG_DIR := $(DEMO_DIR)/engineering

# ============================================================================
# Help
# ============================================================================
.PHONY: help
help: ## Show this help
	@echo "Northstar Group Demo Commands"
	@echo ""
	@echo "=== Golden Path (Start Here) ==="
	@echo ""
	@echo "User Flow:"
	@grep -E '^(setup|verify|run|reset|reset-all|doctor):.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-25s %s\n", $$1, $$2}'
	@echo ""
	@echo "=== Advanced Commands ==="
	@echo ""
	@echo "Scenario Management:"
	@grep -E '^(list-|describe-|validate-|ui-|health).*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-25s %s\n", $$1, $$2}'
	@echo ""
	@echo "SRE Track (Advanced):"
	@grep -E '^sre-.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-25s %s\n", $$1, $$2}'
	@echo ""
	@echo "Engineering Track (Advanced):"
	@grep -E '^eng-.*:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-25s %s\n", $$1, $$2}'

# ============================================================================
# User Flow (Golden Path)
# ============================================================================
.PHONY: setup
setup: ## One-time setup (installs UI testing dependencies)
	@$(SHARED_DIR)/scripts/democtl.sh setup

.PHONY: verify
verify: ## Check all prerequisites (SRE + Engineering + UI)
	@$(SHARED_DIR)/scripts/democtl.sh verify

.PHONY: run
run: ## Run a scenario (auto-detects track, ensures runtime)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && echo "Usage: make run SCENARIO=<track>/<slug>" && exit 1)
	@$(SHARED_DIR)/scripts/democtl.sh run $(SCENARIO)

.PHONY: reset
reset: ## Reset a scenario (down + up)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && echo "Usage: make reset SCENARIO=<track>/<slug>" && exit 1)
	@$(SHARED_DIR)/scripts/democtl.sh reset $(SCENARIO)

.PHONY: reset-all
reset-all: ## Master reset (requires FORCE=true, removes all runtimes)
	@$(SHARED_DIR)/scripts/democtl.sh reset-all

.PHONY: doctor
doctor: ## Show status of all runtimes (non-failing)
	@$(SHARED_DIR)/scripts/democtl.sh doctor

# Convenience aliases for discoverability
.PHONY: sre-run
sre-run: ## (Alias) Run an SRE scenario
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(SHARED_DIR)/scripts/democtl.sh run $(SCENARIO)

.PHONY: eng-run
eng-run: ## (Alias) Run an Engineering scenario
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(SHARED_DIR)/scripts/democtl.sh run $(SCENARIO)

# ============================================================================
# Scenario Management
# ============================================================================
.PHONY: health
health: ## Run health checks on a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@if [ -d "$(SRE_DIR)/scenarios/$(SCENARIO)" ]; then \
		make sre-health SCENARIO=$(SCENARIO); \
	elif [ -d "$(ENG_DIR)/scenarios/$(SCENARIO)" ]; then \
		make eng-health SCENARIO=$(SCENARIO); \
	else \
		echo "Error: Scenario not found: $(SCENARIO)"; exit 1; \
	fi
.PHONY: list-scenarios
list-scenarios: ## List all available scenarios
	@$(SHARED_DIR)/scripts/validate-scenarios.sh list

.PHONY: describe-scenario
describe-scenario: ## Describe a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(SHARED_DIR)/scripts/validate-scenarios.sh describe $(SCENARIO)

.PHONY: validate-scenarios
validate-scenarios: ## Validate all scenario manifests
	@$(SHARED_DIR)/scripts/validate-scenarios.sh validate

# ============================================================================
# SRE Track
# ============================================================================
.PHONY: sre-prereqs
sre-prereqs: ## Check SRE track prerequisites
	@$(SRE_DIR)/scripts/check-prereqs.sh

.PHONY: sre-cluster-up
sre-cluster-up: ## Create Kind cluster for SRE demos
	@echo "Creating Kind cluster..."
	@kind create cluster --config $(SRE_DIR)/kind/cluster.yaml

.PHONY: sre-cluster-down
sre-cluster-down: ## Delete Kind cluster
	@echo "Deleting Kind cluster..."
	@kind delete cluster --name fider-demo

.PHONY: sre-setup
sre-setup: sre-prereqs sre-cluster-up ## Full SRE setup (prereqs + cluster + gateway)
	@echo "Installing Gateway API CRDs..."
	@kubectl --context=$(KUBE_CONTEXT) apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.2.0/standard-install.yaml
	@echo "Installing Envoy Gateway controller..."
	@helm upgrade --install envoy-gateway oci://docker.io/envoyproxy/gateway-helm \
		--kube-context=$(KUBE_CONTEXT) \
		--version v1.2.4 \
		--namespace envoy-gateway-system --create-namespace \
		--skip-crds \
		--wait --timeout 5m
	@echo "Applying shared Gateway resources..."
	@kubectl --context=$(KUBE_CONTEXT) apply -f $(SRE_DIR)/base/gateway.yaml
	@echo "SRE environment ready!"

.PHONY: sre-demo
sre-demo: ## Deploy a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@KUBE_CONTEXT=$(KUBE_CONTEXT) $(SRE_DIR)/scripts/demo.sh up $(SCENARIO)

.PHONY: sre-down
sre-down: ## Remove a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@KUBE_CONTEXT=$(KUBE_CONTEXT) $(SRE_DIR)/scripts/demo.sh down $(SCENARIO)

.PHONY: sre-reset
sre-reset: ## Reset a scenario (down + up)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@KUBE_CONTEXT=$(KUBE_CONTEXT) $(SRE_DIR)/scripts/demo.sh reset $(SCENARIO)

.PHONY: sre-down-all
sre-down-all: ## Remove all demo namespaces
	@KUBE_CONTEXT=$(KUBE_CONTEXT) $(SRE_DIR)/scripts/demo.sh down-all

.PHONY: sre-verify
sre-verify: ## Verify scenario deployment (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@KUBE_CONTEXT=$(KUBE_CONTEXT) $(SHARED_DIR)/scripts/run-checks.sh sre $(SCENARIO) verify $(if $(STAGE),--stage $(STAGE))

.PHONY: sre-health
sre-health: ## Health check a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@KUBE_CONTEXT=$(KUBE_CONTEXT) $(SHARED_DIR)/scripts/run-checks.sh sre $(SCENARIO) health $(if $(STAGE),--stage $(STAGE))

.PHONY: sre-status
sre-status: ## Show cluster status
	@kubectl --context=$(KUBE_CONTEXT) get pods -A | grep -E "(demo-|envoy)" || echo "No demo pods running"

# ============================================================================
# Engineering Track
# ============================================================================
.PHONY: eng-prereqs
eng-prereqs: ## Check Engineering track prerequisites
	@$(ENG_DIR)/scripts/check-prereqs.sh

.PHONY: eng-edge-up
eng-edge-up: ## Start Traefik edge proxy
	@docker compose -f $(ENG_DIR)/compose/edge/docker-compose.yml up -d

.PHONY: eng-edge-down
eng-edge-down: ## Stop Traefik edge proxy
	@docker compose -f $(ENG_DIR)/compose/edge/docker-compose.yml down

.PHONY: eng-testdeps-up
eng-testdeps-up: ## Start test dependencies (postgres, minio)
	@docker compose -f $(ENG_DIR)/compose/testdeps/docker-compose.yml up -d

.PHONY: eng-testdeps-down
eng-testdeps-down: ## Stop test dependencies
	@docker compose -f $(ENG_DIR)/compose/testdeps/docker-compose.yml down -v

.PHONY: eng-scenario-init
eng-scenario-init: ## Initialize scenario worktree (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/worktree.sh init $(SCENARIO)

.PHONY: eng-scenario-reset
eng-scenario-reset: ## Reset scenario worktree (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/worktree.sh reset $(SCENARIO)

.PHONY: eng-scenario-clean
eng-scenario-clean: ## Remove scenario worktree (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/worktree.sh remove $(SCENARIO)

.PHONY: eng-up
eng-up: ## Start a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/demo.sh up $(SCENARIO)

.PHONY: eng-down
eng-down: ## Stop a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/demo.sh down $(SCENARIO)

.PHONY: eng-reset
eng-reset: ## Reset a scenario (down + worktree reset + up)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/demo.sh reset $(SCENARIO)

.PHONY: eng-sniff
eng-sniff: ## Follow scenario logs (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/demo.sh sniff $(SCENARIO)

.PHONY: eng-status
eng-status: ## Show scenario status (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(ENG_DIR)/scripts/demo.sh status $(SCENARIO)

.PHONY: eng-verify
eng-verify: ## Verify scenario deployment (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(SHARED_DIR)/scripts/run-checks.sh engineering $(SCENARIO) verify $(if $(STAGE),--stage $(STAGE))

.PHONY: eng-health
eng-health: ## Health check a scenario (SCENARIO required)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(SHARED_DIR)/scripts/run-checks.sh engineering $(SCENARIO) health $(if $(STAGE),--stage $(STAGE))

.PHONY: eng-setup
eng-setup: eng-prereqs eng-edge-up ## Full Engineering setup (prereqs + edge proxy)
	@echo "Engineering environment ready!"

# ============================================================================
# UI Verification (Cross-track)
# ============================================================================
.PHONY: ui-setup
ui-setup: ## Install Playwright dependencies for UI verification
	@echo "Installing Playwright dependencies..."
	@cd $(DEMO_DIR)/ui && npm ci
	@echo "Installing Playwright browsers..."
	@cd $(DEMO_DIR)/ui && npx playwright install chromium
	@echo "UI verification ready!"

.PHONY: ui-verify
ui-verify: ## Run UI/Playwright checks only (TYPE and SCENARIO required)
	@test -n "$(TYPE)" || (echo "Error: TYPE is required (sre or engineering)" && exit 1)
	@test -n "$(SCENARIO)" || (echo "Error: SCENARIO is required" && exit 1)
	@$(SHARED_DIR)/scripts/run-checks.sh $(TYPE) $(SCENARIO) verify --only playwright $(if $(STAGE),--stage $(STAGE))
