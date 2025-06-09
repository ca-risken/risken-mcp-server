include .env
HTTP_PORT ?= 8098

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

.PHONY: oauth
oauth: build
	@docker run -it --rm \
		-e RISKEN_URL=${RISKEN_URL} \
		-e MCP_SERVER_URL=http://localhost:${HTTP_PORT} \
		-e CLIENT_ID=${CLIENT_ID} \
		-e CLIENT_SECRET=${CLIENT_SECRET} \
		-e AUTHZ_METADATA_ENDPOINT=${AUTHZ_METADATA_ENDPOINT} \
		-e JWT_SIGNING_KEY=${JWT_SIGNING_KEY} \
		-p ${HTTP_PORT}:8080 \
		risken-mcp-server oauth

.PHONY: logs
logs:
	@docker logs -f $$(docker ps -q --filter "ancestor=risken-mcp-server" | head -1)

.PHONY: help
help:
	docker run -it --rm risken-mcp-server help

############################################################
# Google Cloud Run
############################################################
.PHONY: gcp-login
gcp-login:
	@gcloud auth login
	@gcloud auth application-default login

.PHONY: gcp-plan
gcp-plan:
	@cd terraform/examples/test && \
	terraform plan

.PHONY: gcp-apply
gcp-apply:
	@cd terraform/examples/test && \
	terraform apply

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
		-H "RISKEN-ACCESS-TOKEN: ${RISKEN_ACCESS_TOKEN}" \
		-d '{"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}') && \
	SESSION_ID=$$(echo "$$RESPONSE" | grep -i "mcp-session-id:" | sed 's/.*mcp-session-id: *\([^ \r]*\).*/\1/' | tr -d '\r\n') && \
	echo "$$SESSION_ID" > $(MCP_SESSION_FILE)

.PHONY: cat-mcp-session
cat-mcp-session:
	@cat $(MCP_SESSION_FILE)

.PHONY: http-tools-list
http-tools-list: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "RISKEN-ACCESS-TOKEN: ${RISKEN_ACCESS_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' \
		| jq .

.PHONY: http-get-project
http-get-project: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "RISKEN-ACCESS-TOKEN: ${RISKEN_ACCESS_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project"}}' \
		| jq .

.PHONY: http-get-project2
http-get-project2: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "RISKEN-ACCESS-TOKEN: ${RISKEN_ACCESS_TOKEN2}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_project"}}' \
		| jq .

.PHONY: http-search-finding
http-search-finding: http-get-session
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "RISKEN-ACCESS-TOKEN: ${RISKEN_ACCESS_TOKEN}" \
		-H "$$(cat $(MCP_SESSION_FILE))" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_finding","arguments":{"from_score":0.7,"limit":5}}}' \
		| jq .

.PHONY: http-error-no-auth
http-error-no-auth:
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":999,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}' \
		| jq .

.PHONY: http-error-invalid-auth
http-error-invalid-auth:
	@curl -s -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-H "RISKEN-ACCESS-TOKEN: INVALID_TOKEN" \
		-d '{"jsonrpc":"2.0","id":999,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}' \
		| jq .

############################################################
# OAuth MCP Server Requests
############################################################
.PHONY: oauth-mcp-metadata
oauth-mcp-metadata:
	curl -s -X GET http://127.0.0.1:${HTTP_PORT}/.well-known/oauth-protected-resource \
		| jq .

.PHONY: oauth-authz-metadata
oauth-authz-metadata:
	curl -s -X GET http://127.0.0.1:${HTTP_PORT}/.well-known/oauth-authorization-server \
		| jq .

.PHONY: oauth-init
oauth-init:
	@curl -i -X POST http://127.0.0.1:${HTTP_PORT}/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","clientInfo":{"name":"curl-client"}}}'
