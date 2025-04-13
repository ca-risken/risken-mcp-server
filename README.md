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
        "ghcr.io/ca-risken/risken-mcp-server",
        "stdio"
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

### Project

- *get_project* - Get RISKEN project.

### Finding

- *search_finding* - Search RISKEN findings.
  - `finding_id` - Search by finding ID.
  - `alert_id` - Search by alert ID.
  - `data_source` - Search by data source.
  - `resource_name` - Search by resource name.
  - `from_score` - Search by minimum score.
    - `0.0` ~ `0.3` - Low
    - `0.3` ~ `0.6` - Medium
    - `0.6` ~ `0.8` - High
    - `0.8` ~ `1.0` - Critical
  - `status` - Search by status.
    - `0` - All
    - `1` - Active (default)
    - `2` - Pending
  - `offset` - Search by offset.
  - `limit` - Search by limit.

- *archive_finding* - Archive RISKEN finding.
  - `finding_id` - Archive by finding ID.
  - `note` - Note.

### Alert

- *search_alert* - Search RISKEN alert.
  - `status` - Search by status.
    - `1` - Active
    - `2` - Pending
    - `3` - Deactive (already closed)

## Resources

### Finding Contents

- *Get Finding Contents* Retrieves the content of a specific finding.
  - *Template*: `finding://{project_id}/{finding_id}`
  - *Parameters*:
    - `project_id`: The ID of the project.
    - `finding_id`: The ID of the finding.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
