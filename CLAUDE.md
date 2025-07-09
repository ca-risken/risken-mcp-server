# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RISKEN MCP Server is a Model Context Protocol server that provides integration with RISKEN security APIs. It offers tools for security finding management, alert monitoring, and project access through MCP-compliant clients like Claude Desktop.

## Build and Development Commands

### Docker Commands
- `make build` - Build Docker image locally
- `make stdio` - Run server in stdio mode with Docker (requires RISKEN_URL and RISKEN_ACCESS_TOKEN in .env)
- `make http` - Run server in HTTP mode on port 8098 (or HTTP_PORT env var)
- `make oauth` - Run server with OAuth2.1 authentication
- `make logs` - View logs from running Docker container
- `make help` - Show help from Docker container

### Testing
- `go test ./...` - Run all Go tests
- Test files are located in `pkg/helper/*_test.go` and `pkg/riskenmcp/*_test.go`

### Example API Calls
- `make stdio-get-project` - Test stdio mode with get_project tool
- `make http-health` - Test HTTP health endpoint
- `make http-get-project` - Test HTTP mode with authentication
- `make http-search-finding` - Test finding search functionality

### Google Cloud Deployment
- `make gcp-login` - Authenticate with Google Cloud
- `make gcp-plan` - Plan Terraform deployment
- `make gcp-apply` - Apply Terraform deployment

## Architecture

### Core Components

**Entry Points (`cmd/risken-mcp-server/`)**:
- `main.go` - CLI application using Cobra framework
- `stdio.go` - Standard I/O MCP server implementation  
- `http.go` - HTTP MCP server implementation
- `oauth.go` - OAuth2.1 authentication server

**MCP Server (`pkg/riskenmcp/`)**:
- `server.go` - Main MCP server factory and configuration
- `project.go`, `finding.go`, `alert.go` - Tool implementations for RISKEN API operations
- `finding_search.go`, `finding_archive.go` - Specific finding operations
- `context.go` - Request context and client management

**Authentication (`pkg/oauth/`)**:
- Full OAuth2.1 implementation with Third-Party Authorization Flow
- JWT token validation and session management
- Dynamic client registration support

**Utilities (`pkg/helper/`)**:
- `risken.go` - RISKEN API client wrapper
- `request.go` - HTTP request helpers with authentication
- `json.go` - JSON handling utilities

### Server Modes

1. **Stdio Mode**: Direct MCP communication via stdin/stdout
2. **HTTP Mode**: Streamable HTTP with session management
3. **OAuth Mode**: HTTP with OAuth2.1 authentication flow

### MCP Tools Available

- `get_project` - Retrieve RISKEN project information
- `search_finding` - Search security findings with filters (score, status, data source)
- `archive_finding` - Archive resolved findings
- `search_alert` - Search active alerts

### Resources

- `finding://{project_id}/{finding_id}` - Access specific finding content

## Environment Configuration

Required environment variables are defined in `.env`:
- `RISKEN_URL` - RISKEN API endpoint
- `RISKEN_ACCESS_TOKEN` - Authentication token

For OAuth mode, additional variables:
- `MCP_SERVER_URL` - Public URL of MCP server
- `CLIENT_ID`, `CLIENT_SECRET` - OAuth client credentials  
- `AUTHZ_METADATA_ENDPOINT` - Identity provider metadata
- `JWT_SIGNING_KEY` - JWT signing secret

## OAuth 2.1 Architecture

### Three-Party OAuth Flow

RISKEN MCP Server implements OAuth 2.1 Third-Party Authorization Flow with three distinct entities:

1. **MCP Client** (OAuth 2.1 Client)
   - Implements PKCE (code_challenge/code_verifier)
   - Initiates authorization requests to MCP Server

2. **RISKEN MCP Server** (Dual Role)
   - **Authorization Server**: Issues tokens to MCP Clients after PKCE verification
   - **Resource Server**: Protects RISKEN API access using issued tokens

3. **External IdP** (Identity Provider)
   - Performs user authentication (Auth0, Keycloak, etc.)
   - Issues tokens for user identity verification

### Authorization Flow Steps

```
1. MCP Client → MCP Server/authorize: Start OAuth with PKCE challenge
2. MCP Server → External IdP: Redirect user for authentication  
3. User → External IdP: Login authentication
4. External IdP → MCP Server/callback: Return authorization code
5. MCP Server: Store IdP code, generate internal JWT authorization code
6. MCP Server → MCP Client: Redirect with internal authorization code
7. MCP Client → MCP Server/token: Exchange code + PKCE verifier
8. MCP Server: Verify PKCE → Exchange IdP code → Return access token
```

### Key Implementation Details

**Delayed Token Exchange**: IdP authorization code is stored at callback but only exchanged for access token after MCP Client's PKCE verification succeeds. This ensures only legitimate MCP Clients can trigger IdP resource consumption.

**JWT Authorization Codes**: Internal 5-minute JWT tokens contain encrypted session data (PKCE challenge, IdP code, redirect URI) to maintain stateless operation.

**PKCE Verification**: Performed between MCP Client ↔ MCP Server to prevent authorization code interception attacks, independent of IdP capabilities.

## Development Notes

- Uses Go 1.23.3 with `github.com/mark3labs/mcp-go` for MCP protocol
- All API interactions go through `github.com/ca-risken/go-risken` client
- Authentication supports both direct tokens and OAuth2.1 flows
- Terraform modules provided for Google Cloud Run deployment in `terraform/examples/`

## Security Best Practices

- Always validate environment variables are set before starting server
- Implement proper input validation for all MCP tool parameters
- Add security headers (CORS, CSP, HSTS) to HTTP responses
- Use structured logging to avoid accidental secret exposure
- Consider implementing token refresh mechanisms for longer sessions