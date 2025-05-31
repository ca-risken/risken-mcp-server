include .env
HTTP_PORT ?= 8080
MCP_AUTH_TOKEN ?= xxxxxx

.PHONY: build
build:
	docker build -t risken-mcp-server .

.PHONY: stdio
stdio: build
	@docker run -it --rm \
		-e RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} \
		-e RISKEN_URL=${RISKEN_URL} \
		risken-mcp-server stdio

.PHONY: http
http: build
	@docker run -it --rm \
		-e RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} \
		-e RISKEN_URL=${RISKEN_URL} \
		-e MCP_AUTH_TOKEN=${MCP_AUTH_TOKEN} \
		-p ${HTTP_PORT}:8080 \
		risken-mcp-server http

.PHONY: logs
logs:
	@docker logs -f $$(docker ps -q --filter "ancestor=risken-mcp-server" | head -1)

.PHONY: help
help:
	docker run -it --rm risken-mcp-server help

############################################################
# Stdio MCP Server
############################################################
.PHONY: call-get-project-stdio
call-get-project-stdio:
	@export RISKEN_ACCESS_TOKEN=${RISKEN_ACCESS_TOKEN} && \
	export RISKEN_URL=${RISKEN_URL} && \
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
# Streamable HTTP MCP Server
############################################################
MCP_SESSION_FILE := /tmp/risken-mcp-session-id

.PHONY: get-mcp-session
get-mcp-session:
	@RESPONSE=$$(curl -s -i -XPOST \
	http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${MCP_AUTH_TOKEN}" \
		-d '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}') && \
	SESSION_ID=$$(echo "$$RESPONSE" | grep -i "mcp-session-id:" | sed 's/.*mcp-session-id: *\([^ \r]*\).*/\1/' | tr -d '\r\n') && \
	echo "$$SESSION_ID" > $(MCP_SESSION_FILE)

.PHONY: call-get-project-http
call-get-project-http: get-mcp-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${MCP_AUTH_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project"}}' \
		| jq .

.PHONY: call-search-finding-http
call-search-finding-http: get-mcp-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer ${MCP_AUTH_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_finding","arguments":{"from_score":0.7,"limit":5}}}' \
		| jq .

.PHONY: call-invalid-http
call-invalid-http: get-mcp-session
	@curl -i -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "Authorization: Bearer INVALID_TOKEN" \
		-d '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}'
