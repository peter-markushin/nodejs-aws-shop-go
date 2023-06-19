# See https://tech.davis-hansson.com/p/make/
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.DEFAULT_GOAL := help
.PHONY: help
help:
	@printf "\033[33mUsage:\033[0m\n  make TARGET\n\033[33m\nAvailable Commands:\n\033[0m"
	@grep -E '^[a-zA-Z-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  [32m%-27s[0m %s\n", $$1, $$2}'

.PHONY: build
build:
	./build-handlers.sh

.PHONY: deploy
deploy: build
	if [ -z "$(email)" ]; then echo "Email is required"; exit 1; fi
	cd infra; cdk deploy --parameters NotificationEmail=$(email)

