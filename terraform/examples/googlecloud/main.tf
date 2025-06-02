terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 4.84.0"
    }
  }
}

provider "google" {
  project = "your-project-id"
  region  = "your-region"
}

module "risken_mcp_server" {
  source = "github.com/ca-risken/risken-mcp-server//terraform/modules/cloudrun?ref=main"

  project_id   = "your-project-id"
  region       = "your-region"
  service_name = "risken-mcp-server"
  risken_url   = "https://your-risken-url"
}

output "mcp_endpoint" {
  description = "MCP endpoint URL"
  value       = module.risken_mcp_server.mcp_endpoint
}
