# infrastructure/terraform/modules/tenant-provisioning/main.tf
# Multi-tenant SaaS provisioning automation

variable "tenant_id" {
  description = "Unique tenant identifier"
  type        = string
}

variable "tenant_tier" {
  description = "Tenant subscription tier"
  type        = string
  validation {
    condition     = contains(["standard", "enterprise", "sovereign"], var.tenant_tier)
    error_message = "Tier must be standard, enterprise, or sovereign."
  }
}

variable "region" {
  description = "Primary deployment region"
  type        = string
}

# Tenant namespace with resource quotas
resource "kubernetes_namespace" "tenant" {
  metadata {
    name = "tenant-${var.tenant_id}"
    labels = {
      "nexus-platform.io/tenant-id"   = var.tenant_id
      "nexus-platform.io/tenant-tier" = var.tenant_tier
    }
  }
}

resource "kubernetes_resource_quota" "tenant_quota" {
  metadata {
    name      = "tenant-${var.tenant_id}-quota"
    namespace = kubernetes_namespace.tenant.metadata[0].name
  }
  spec {
    hard = {
      "limits.cpu"       = var.tenant_tier == "enterprise" ? "100" : var.tenant_tier == "sovereign" ? "500" : "10"
      "limits.memory"    = var.tenant_tier == "enterprise" ? "200Gi" : var.tenant_tier == "sovereign" ? "1Ti" : "20Gi"
      "pods"             = var.tenant_tier == "enterprise" ? "1000" : var.tenant_tier == "sovereign" ? "5000" : "100"
    }
  }
}

resource "kubernetes_network_policy" "tenant_isolation" {
  metadata {
    name      = "tenant-${var.tenant_id}-isolation"
    namespace = kubernetes_namespace.tenant.metadata[0].name
  }
  spec {
    pod_selector {}
    policy_types = ["Ingress", "Egress"]
    ingress {
      from {
        namespace_selector {
          match_labels = {
            "nexus-platform.io/tenant-id" = var.tenant_id
          }
        }
      }
    }
  }
}

output "tenant_namespace" {
  value = kubernetes_namespace.tenant.metadata[0].name
}
