
resource "google_compute_global_address" "lb" {
  name = "lb-ip"
}

provider "google-beta" {
  project = var.project
}

module "lb-http" {
  depends_on = [
    google_project_service.compute
  ]

  source  = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version = "~> 5.1"

  project = var.project
  name    = "lb"

  ssl                             = true
  managed_ssl_certificate_domains = ["grpco.in", "api.grpco.in"]
  https_redirect                  = true
  address                         = google_compute_global_address.lb.address

  backends = {
    default = {
      description            = "web frontend server"
      timeout_sec            = 30
      enable_cdn             = false
      custom_request_headers = null
      security_policy        = null
      log_config = {
        enable      = true
        sample_rate = 1.0
      }
      groups = [
        {
          group = google_compute_region_network_endpoint_group.frontend.id
        }
      ]
      iap_config = {
        enable               = false
        oauth2_client_id     = null
        oauth2_client_secret = null
      }
    }
    api = {
      description            = "api server"
      timeout_sec            = 86400 // it's longer than Cloud Run timeout anyway
      enable_cdn             = false
      custom_request_headers = null
      security_policy        = null
      log_config = {
        enable      = true
        sample_rate = 1.0
      }
      groups = [
        {
          group = google_compute_region_network_endpoint_group.apiserver.id
        }
      ]
      iap_config = {
        enable               = false
        oauth2_client_id     = null
        oauth2_client_secret = null
      }
    }
  }
  create_url_map = false
  url_map        = google_compute_url_map.default.id
}


resource "google_compute_region_network_endpoint_group" "frontend" {
  depends_on = [
    google_project_service.compute
  ]
  project = var.project
  region  = var.region

  name                  = "cr-fe-neg"
  network_endpoint_type = "SERVERLESS"
  cloud_run {
    service = google_cloud_run_service.frontend.name
  }
}

resource "google_compute_region_network_endpoint_group" "apiserver" {
  depends_on = [
    google_project_service.compute
  ]
  project = var.project
  region  = var.region

  name                  = "cr-api-neg"
  network_endpoint_type = "SERVERLESS"
  cloud_run {
    service = google_cloud_run_service.apiserver.name
  }
}

resource "google_compute_url_map" "default" {
  depends_on = [
    google_project_service.compute
  ]
  project         = var.project
  name            = "urlmap"
  default_service = module.lb-http.backend_services["default"].self_link
  host_rule {
    hosts        = ["api.grpco.in"]
    path_matcher = "api"
  }
  path_matcher {
    name            = "api"
    default_service = module.lb-http.backend_services["api"].self_link
  }
}
