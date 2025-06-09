output "service_url" {
  description = "URL of the Cloud Run service"
  value       = google_cloud_run_service.risken_mcp_server.status[0].url
}

output "service_name" {
  description = "Name of the Cloud Run service"
  value       = google_cloud_run_service.risken_mcp_server.name
}

output "service_account_email" {
  description = "Email of the Cloud Run service account"
  value       = google_service_account.cloud_run.email
}

output "artifact_registry_url" {
  description = "Artifact Registry repository URL"
  value       = var.create_artifact_registry ? google_artifact_registry_repository.risken_mcp[0].name : null
}

output "mcp_endpoint" {
  description = "MCP endpoint URL"
  value       = "${google_cloud_run_service.risken_mcp_server.status[0].url}/mcp"
}

