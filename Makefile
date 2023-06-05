# See https://tech.davis-hansson.com/p/make/
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

.DEFAULT_GOAL := help
.PHONY: help
help:
	@printf "\033[33mUsage:\033[0m\n  make TARGET\n\033[33m\nAvailable Commands:\n\033[0m"
	@grep -E '^[a-zA-Z-]+:.*?## .*$$' Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  [32m%-27s[0m %s\n", $$1, $$2}'

build ./tmp/lambdaHandler:
	cd app; GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ../tmp/lambdaHandler .
	#cd tmp; zip -q lambda-handler.zip lambdaHandler

run:
	cd app; go run main.go

deploy: ./tmp/lambdaHandler
	cd infra; cdk deploy
