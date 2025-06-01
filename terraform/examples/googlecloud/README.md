# Terraform for Cloud Run

## Prerequisites

- Terraform
- Google Cloud SDK
- Google Cloud Project
- Docker

## Usage

### Authenticate to Google Cloud

```bash
gcloud auth application-default login
gcloud config set project ca-security-hub
```

### Apply Terraform

```bash
terraform init
terraform plan
terraform apply
```

### Check MCP endpoint

```bash
terraform output
> mcp_endpoint = "https://your-mcp-server.run.app/mcp"
```

