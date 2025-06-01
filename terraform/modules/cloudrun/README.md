# Terraform Module for Cloud Run

You can deploy a **Remote MCP server** on Google Cloud Run.

## Usage

See [terraform/examples/googlecloud](../../examples/googlecloud) for example usage.


<!-- BEGIN_TF_DOCS -->
## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0 |
| <a name="requirement_google"></a> [google](#requirement\_google) | >= 4.84.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_external"></a> [external](#provider\_external) | n/a |
| <a name="provider_google"></a> [google](#provider\_google) | >= 4.84.0 |
| <a name="provider_null"></a> [null](#provider\_null) | n/a |

## Modules

No modules.

## Resources

| Name | Type |
|------|------|
| [google_artifact_registry_repository.risken_mcp](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/artifact_registry_repository) | resource |
| [google_artifact_registry_repository_iam_member.cloud_run_artifact_registry](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/artifact_registry_repository_iam_member) | resource |
| [google_cloud_run_service.risken_mcp_server](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_service) | resource |
| [google_cloud_run_service_iam_policy.noauth](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_service_iam_policy) | resource |
| [google_secret_manager_secret.risken_access_token](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret) | resource |
| [google_secret_manager_secret_iam_member.cloud_run_secret_access](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret_iam_member) | resource |
| [google_secret_manager_secret_version.risken_access_token](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/secret_manager_secret_version) | resource |
| [google_service_account.cloud_run](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account) | resource |
| [null_resource.copy_ghcr_image](https://registry.terraform.io/providers/hashicorp/null/latest/docs/resources/resource) | resource |
| [external_external.ghcr_image_digest](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/external) | data source |
| [google_iam_policy.noauth](https://registry.terraform.io/providers/hashicorp/google/latest/docs/data-sources/iam_policy) | data source |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_artifact_registry_name"></a> [artifact\_registry\_name](#input\_artifact\_registry\_name) | Artifact Registry repository name | `string` | `"risken-mcp"` | no |
| <a name="input_cpu_limit"></a> [cpu\_limit](#input\_cpu\_limit) | CPU limit | `string` | `"0.5"` | no |
| <a name="input_create_artifact_registry"></a> [create\_artifact\_registry](#input\_create\_artifact\_registry) | Create Artifact Registry repository | `bool` | `true` | no |
| <a name="input_deletion_protection"></a> [deletion\_protection](#input\_deletion\_protection) | Enable deletion protection | `bool` | `false` | no |
| <a name="input_image_retention_count"></a> [image\_retention\_count](#input\_image\_retention\_count) | Number of images to retain in Artifact Registry | `number` | `10` | no |
| <a name="input_max_instances"></a> [max\_instances](#input\_max\_instances) | Maximum number of instances | `number` | `10` | no |
| <a name="input_memory_limit"></a> [memory\_limit](#input\_memory\_limit) | Memory limit | `string` | `"512Mi"` | no |
| <a name="input_min_instances"></a> [min\_instances](#input\_min\_instances) | Minimum number of instances | `number` | `1` | no |
| <a name="input_project_id"></a> [project\_id](#input\_project\_id) | Google Cloud project ID | `string` | n/a | yes |
| <a name="input_region"></a> [region](#input\_region) | Google Cloud region | `string` | `"asia-northeast1"` | no |
| <a name="input_risken_access_token"></a> [risken\_access\_token](#input\_risken\_access\_token) | RISKEN access token | `string` | n/a | yes |
| <a name="input_risken_url"></a> [risken\_url](#input\_risken\_url) | RISKEN URL | `string` | n/a | yes |
| <a name="input_service_name"></a> [service\_name](#input\_service\_name) | Cloud Run service name | `string` | `"risken-mcp-server"` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_artifact_registry_url"></a> [artifact\_registry\_url](#output\_artifact\_registry\_url) | Artifact Registry repository URL |
| <a name="output_mcp_client_config"></a> [mcp\_client\_config](#output\_mcp\_client\_config) | MCP client configuration |
| <a name="output_mcp_endpoint"></a> [mcp\_endpoint](#output\_mcp\_endpoint) | MCP endpoint URL |
| <a name="output_service_name"></a> [service\_name](#output\_service\_name) | Cloud Run service name |
| <a name="output_service_url"></a> [service\_url](#output\_service\_url) | Cloud Run service URL |
<!-- END_TF_DOCS -->