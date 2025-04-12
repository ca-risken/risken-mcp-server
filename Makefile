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
