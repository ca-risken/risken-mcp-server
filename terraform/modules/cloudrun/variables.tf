variable "project_id" {
  description = "Google Cloud project ID"
  type        = string
}

variable "region" {
  description = "Google Cloud region"
  type        = string
  default     = "asia-northeast1"
}

variable "service_name" {
  description = "Cloud Run service name"
  type        = string
  default     = "risken-mcp-server"
}

variable "risken_url" {
  description = "RISKEN URL"
  type        = string
}

# OAuth Configuration
variable "mcp_server_url" {
  description = "Public URL of MCP server"
  type        = string
}

variable "authz_metadata_endpoint" {
  description = "IdP's OAuth metadata endpoint"
  type        = string
}

# Client ID Configuration (choose one)
variable "client_id_name" {
  description = "Name of existing Secret Manager secret containing OAuth client ID"
  type        = string
  default     = null
}

# Secret Manager References (must be pre-created)
variable "client_secret_name" {
  description = "Name of existing Secret Manager secret containing OAuth client secret"
  type        = string
}

variable "jwt_signing_key_name" {
  description = "Name of existing Secret Manager secret containing JWT signing key"
  type        = string
}

variable "cpu_limit" {
  description = "CPU limit"
  type        = string
  default     = "0.5"
}

variable "memory_limit" {
  description = "Memory limit"
  type        = string
  default     = "512Mi"
}

variable "min_instances" {
  description = "Minimum number of instances"
  type        = number
  default     = 1
}

variable "max_instances" {
  description = "Maximum number of instances"
  type        = number
  default     = 10
}

variable "create_artifact_registry" {
  description = "Create Artifact Registry repository"
  type        = bool
  default     = true
}

variable "artifact_registry_name" {
  description = "Artifact Registry repository name"
  type        = string
  default     = "risken-mcp"
}

variable "image_retention_count" {
  description = "Number of images to retain in Artifact Registry"
  type        = number
  default     = 10
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = false
}

variable "debug" {
  description = "Enable debug logging"
  type        = bool
  default     = false
}
