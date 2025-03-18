# Makefile to run tests

# Variables
RUN_INITIAL_BUILD_SCRIPT = build.sh

# Colors
BLUE := \033[0;34m
YELLOW := \033[0;33m
GREEN := \033[0;32m
RED := \033[0;31m
RESET := \033[0m

# Progress function
define run_with_progress
	@printf "###############################################################################################################################################\n"
	@printf "$(BLUE)⏳ Task: \"$(1)\"$(RESET)\n"
	@echo
	$(eval START_TIME := $(shell date +%s))
	@$(2)
	@printf "$(GREEN)✅ Task Complete: \"$(1)\"$(RESET)\n"
	@echo
	@printf "$(GREEN)⏱️  Time taken: $$(($$(date +%s) - $(START_TIME))) seconds$(RESET)\n"
endef

# Default target
all: run_build_checks run_initial_build_script 

# Run tests
run_build_checks:
	@printf "$(BLUE)🚀 Running build checks$(RESET)\n"
	$(call run_with_progress,'Checking for updates and dependencies',\
		go mod tidy && \
		if go list -m -u all | grep -v '^module' | grep -B 2 Update > /dev/null 2>&1; then \
			printf "$(YELLOW)⚡️ Updates available. Upgrading Go modules$(RESET)\n" && \
			go get -u ./...; \
		else \
			printf "$(GREEN)✅ No updates available. Skipping upgrade$(RESET)\n\n"; \
		fi)

# Run initial build script
run_initial_build_script:
	@echo ;
	@chmod +x "$(RUN_INITIAL_BUILD_SCRIPT)"
	$(call run_with_progress,'Running build script', "./$(RUN_INITIAL_BUILD_SCRIPT)");
	@echo ;

# Clean up generated files
clean:
	@printf "$(BLUE)🧹 Cleaning up$(RESET)\n\n"
	$(call run_with_progress,'Cleaning build artifacts',rm -f *.log)

# Help command
help:
	@printf "$(BLUE)📖 Makefile targets:$(RESET)\n\n"
	@printf "  🎯 all                 - Run static and security tests\n"
	@printf "  🔍 run_build_checks    - Run build checks\n"
	@printf "  🧹 clean              - Clean up generated files\n"
	@printf "  💡 help               - Show this help message\n"

.PHONY: all run_build_checks clean docker-build docker-push help

