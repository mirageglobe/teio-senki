.DEFAULT_GOAL := help

.PHONY: help install run lint test clean

help:                ## Show this help menu
	@grep -E '^[a-zA-Z_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*##"}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

install:             ## Install dependencies with uv
	uv sync

run:                 ## Launch the game
	uv run python main.py

lint:                ## Lint with ruff
	uv run ruff check .

test:                ## Run tests
	uv run pytest

reset:               ## Delete the ledger database (forces reseed on next run)
	rm -f ledger.db

clean:               ## Remove cache and build artefacts
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
