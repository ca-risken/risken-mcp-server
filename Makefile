include .env

.PHONY: build
build:
	docker build -t risken-mcp-server .

.PHONY: run
run:
	@docker run -it --rm \
		-e RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} \
		-e RISKEN_URL=${RISKEN_URL} \
		risken-mcp-server stdio

.PHONY: help
help:
	docker run -it --rm risken-mcp-server help

.PHONY: get-project
get-project:
	@export RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} && \
	export RISKEN_URL=${RISKEN_URL} && \
	echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_project"}}' | \
		go run cmd/risken-mcp-server/main.go stdio  | \
		jq .
