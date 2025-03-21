# Makefile to run tests

# Variables
RUN_INITIAL_LINT_SCRIPT = lint.sh

# Colors
RESET  := $(shell tput sgr0)
BLUE   := $(shell tput setaf 4)
GREEN  := $(shell tput setaf 2)
YELLOW := $(shell tput setaf 3)
RED    := $(shell tput setaf 1)
BOLD   := $(shell tput bold)

# Get terminal width dynamically
TERM_WIDTH := $(shell tput cols)

# Define run_with_progress function with dynamic width progress bar
define run_with_progress
	@TERM_WIDTH=$$(tput cols); \
	PROGRESS_BAR=$$(printf "%*s" "$$TERM_WIDTH" | tr ' ' '='); \
	printf "$$PROGRESS_BAR\n"
	@printf "$(BLUE)$(BOLD)⏳ Task: $(1)$(RESET)\n"
	@echo
	$(eval START_TIME := $(shell date +%s))
	@$(2)
	@printf "$(GREEN)$(BOLD)✅ Task Complete: $(1)$(RESET)\n"
	@echo
	@printf "$(GREEN)⏱️  Time taken: $$(($$(date +%s) - $(START_TIME))) seconds$(RESET)\n"
	@# @TERM_WIDTH=$$(tput cols); \
	@# PROGRESS_BAR=$$(printf "%*s" "$$TERM_WIDTH" | tr ' ' '='); \
	@# printf "$$PROGRESS_BAR\n"
endef

# Default target
all: run_build_checks run_initial_build_script 

# Run tests
run_build_checks:
	@printf "$(BLUE)🚀 Running build checks$(RESET)\n"
	$(call run_with_progress,Checking for updates and dependencies,\
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
	@chmod +x "$(RUN_INITIAL_LINT_SCRIPT)"
	$(call run_with_progress,Running linter and security checks, "./$(RUN_INITIAL_LINT_SCRIPT)");
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

.PHONY: all run_build_checks clean help

