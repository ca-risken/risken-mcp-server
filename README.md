# [WIP] RISKEN MCP Server

The RISKEN MCP Server is a [Model Context Protocol (MCP)](https://modelcontextprotocol.io/introduction) server that provides seamless integration with RISKEN APIs, enabling advanced automation and interaction capabilities for developers and tools.

## Use Cases

- Automating RISKEN data fetching and scanning.
- Extracting and analyzing data from RISKEN.
- Building AI powered tools and applications that interact with RISKEN's ecosystem.

## Prerequisites

1. To run the server in a container, you will need to have [Docker](https://www.docker.com/) installed.
2. Once Docker is installed, you will also need to ensure Docker is running.
3. You will also need to have a [RISKEN Access Token](https://docs.security-hub.jp/en/risken/access_token/).

## Installation

### Claude Desktop

wip

### Cursor

```json
{
  "mcpServers": {
    "risken": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "RISKEN_ACCESS_TOKEN",
        "-e",
        "RISKEN_URL",
        "ghcr.io/ca-risken/risken-mcp-server"
      ],
      "env": {
        "RISKEN_ACCESS_TOKEN": "your_access_token",
        "RISKEN_URL": "http://localhost:8000"
      }
    }
  }
}
```

## Tools

## Resources

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
