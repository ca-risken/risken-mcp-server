output "service_url" {
  description = "Cloud Run service URL"
  value       = google_cloud_run_service.risken_mcp_server.status[0].url
}

output "service_name" {
  description = "Cloud Run service name"
  value       = google_cloud_run_service.risken_mcp_server.name
}

output "artifact_registry_url" {
  description = "Artifact Registry repository URL"
  value       = var.create_artifact_registry ? google_artifact_registry_repository.risken_mcp[0].name : null
}

output "mcp_endpoint" {
  description = "MCP endpoint URL"
  value       = "${google_cloud_run_service.risken_mcp_server.status[0].url}/mcp"
}

