locals {
  target_image_url = "${var.region}-docker.pkg.dev/${var.project_id}/${var.artifact_registry_name}/risken-mcp-server"
  target_image_tag = replace(data.external.ghcr_image_digest.result.digest, "sha256:", "")
}

# Artifact Registry
resource "google_artifact_registry_repository" "risken_mcp" {
  count         = var.create_artifact_registry ? 1 : 0
  location      = var.region
  repository_id = var.artifact_registry_name
  description   = "RISKEN MCP Server container images"
  format        = "DOCKER"

  cleanup_policies {
    id     = "keep-minimum-versions"
    action = "KEEP"
    most_recent_versions {
      keep_count = var.image_retention_count
    }
  }
}

data "external" "ghcr_image_digest" {
  program = ["bash", "-c", <<-EOF
    DIGEST=$(docker manifest inspect ghcr.io/ca-risken/risken-mcp-server:latest 2>/dev/null | \
      jq -r '.manifests[] | select(.platform.architecture=="amd64" and .platform.os=="linux") | .digest' 2>/dev/null || echo "unknown")
    echo "{\"digest\":\"$DIGEST\"}"
EOF
  ]
}

# Copy image from GitHub Container Registry to Artifact Registry
resource "null_resource" "copy_ghcr_image" {
  count = var.create_artifact_registry ? 1 : 0

  triggers = {
    source_image     = "ghcr.io/ca-risken/risken-mcp-server:latest"
    target_image     = local.target_image_url
    target_image_tag = local.target_image_tag
    project_id       = var.project_id
    repository_name  = google_artifact_registry_repository.risken_mcp[0].name
  }
  # Authenticate to Artifact Registry
  provisioner "local-exec" {
    command = "gcloud auth configure-docker ${var.region}-docker.pkg.dev --quiet"
  }
  # Pull image from GitHub Container Registry
  provisioner "local-exec" {
    command = "docker pull --platform linux/amd64 ghcr.io/ca-risken/risken-mcp-server:latest"
  }
  # Tag image
  provisioner "local-exec" {
    command = "docker tag ghcr.io/ca-risken/risken-mcp-server:latest ${local.target_image_url}:${local.target_image_tag}"
  }
  # Push image to Artifact Registry
  provisioner "local-exec" {
    command = "docker push ${local.target_image_url}:${local.target_image_tag}"
  }
  depends_on = [google_artifact_registry_repository.risken_mcp]
}

# Cloud Run v1 API
resource "google_cloud_run_service" "risken_mcp_server" {
  name     = var.service_name
  location = var.region
  project  = var.project_id

  template {
    metadata {
      annotations = {
        # https://cloud.google.com/run/docs/reference/rest/v1/RevisionTemplate
        "autoscaling.knative.dev/minScale" = tostring(var.min_instances)
        "autoscaling.knative.dev/maxScale" = tostring(var.max_instances)
      }
    }

    spec {
      service_account_name = google_service_account.cloud_run.email
      timeout_seconds      = 300 # 5 minutes

      containers {
        image = "${local.target_image_url}:${local.target_image_tag}"
        args  = ["http"]

        ports {
          container_port = 8080
          name           = "http1" # HTTP/1
        }
        env {
          name  = "RISKEN_URL"
          value = var.risken_url
        }

        resources {
          limits = {
            cpu    = var.cpu_limit
            memory = var.memory_limit
          }
        }

        startup_probe {
          http_get {
            path = "/health"
            port = 8080
          }
          initial_delay_seconds = 3
          timeout_seconds       = 30
          period_seconds        = 30
          failure_threshold     = 3
        }
        liveness_probe {
          http_get {
            path = "/health"
            port = 8080
          }
          initial_delay_seconds = 3
          timeout_seconds       = 30
          period_seconds        = 30
          failure_threshold     = 3
        }
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
  depends_on = [null_resource.copy_ghcr_image]
}

data "google_iam_policy" "noauth" {
  binding {
    role = "roles/run.invoker"
    members = [
      "allUsers",
    ]
  }
}

# Service Account
resource "google_service_account" "cloud_run" {
  project      = var.project_id
  account_id   = "${var.service_name}-runner"
  display_name = "Service Account for ${var.service_name} Cloud Run"
}

resource "google_artifact_registry_repository_iam_member" "cloud_run_artifact_registry" {
  count      = var.create_artifact_registry ? 1 : 0
  project    = var.project_id
  location   = var.region
  repository = google_artifact_registry_repository.risken_mcp[0].name
  role       = "roles/artifactregistry.reader"
  member     = "serviceAccount:${google_service_account.cloud_run.email}"
}

resource "google_cloud_run_service_iam_policy" "noauth" {
  location    = google_cloud_run_service.risken_mcp_server.location
  project     = google_cloud_run_service.risken_mcp_server.project
  service     = google_cloud_run_service.risken_mcp_server.name
  policy_data = data.google_iam_policy.noauth.policy_data
}
