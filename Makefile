include .env
HTTP_PORT ?= 8080

.PHONY: generate-docs
generate-docs:
	@hack/generate-tf-docs.sh

.PHONY: build
build:
	docker build -t risken-mcp-server .

.PHONY: stdio
stdio: build
	@docker run -it --rm \
		-e RISKEN_URL=${RISKEN_URL} \
		-e RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} \
		risken-mcp-server stdio

.PHONY: http
http: build
	@docker run -it --rm \
		-e RISKEN_URL=${RISKEN_URL} \
		-p ${HTTP_PORT}:8080 \
		risken-mcp-server http

.PHONY: logs
logs:
	@docker logs -f $$(docker ps -q --filter "ancestor=risken-mcp-server" | head -1)

.PHONY: help
help:
	docker run -it --rm risken-mcp-server help

############################################################
# Stdio MCP Server Requests
############################################################
.PHONY: stdio-get-project
stdio-get-project:
	@export RISKEN_URL=${RISKEN_URL} && \
	export RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} && \
	echo '{ \
		"jsonrpc": "2.0", \
		"id": 1, \
		"method": "tools/call", \
		"params": { \
			"name": "get_project" \
		} \
	}' | \
		go run cmd/risken-mcp-server/*.go stdio | \
		jq .

############################################################
# Streamable HTTP MCP Server Requests
############################################################
.PHONY: http-health
http-health:
	@curl -s -i http://127.0.0.1:${HTTP_PORT}/health

MCP_SESSION_FILE := /tmp/risken-mcp-session-id

.PHONY: http-get-session
http-get-session:
	@RESPONSE=$$(curl -s -i -XPOST \
	http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${RISKEN_ACCESS_TOKEN}" \
		-d '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}') && \
	SESSION_ID=$$(echo "$$RESPONSE" | grep -i "mcp-session-id:" | sed 's/.*mcp-session-id: *\([^ \r]*\).*/\1/' | tr -d '\r\n') && \
	echo "$$SESSION_ID" > $(MCP_SESSION_FILE)

.PHONY: http-tools-list
http-tools-list: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${RISKEN_ACCESS_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'

.PHONY: http-get-project
http-get-project: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${RISKEN_ACCESS_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project"}}' \
		| jq .

.PHONY: http-get-project2
http-get-project2: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${RISKEN_ACCESS_TOKEN2}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project"}}' \
		| jq .

.PHONY: http-search-finding
http-search-finding: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${RISKEN_ACCESS_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_finding","arguments":{"from_score":0.7,"limit":5}}}' \
		| jq .

.PHONY: http-auth-error
http-auth-error:
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer INVALID_TOKEN" \
		-d '{"jsonrpc":"2.0","id":999,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}' \
		| jq .
