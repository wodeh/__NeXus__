terraform {
  required_providers {
    vault = {
      source  = "hashicorp/vault"
      version = "~> 3.20"
    }
  }
}

# ============================================================
# ROOT CA SETUP
# ============================================================

resource "vault_mount" "root_ca" {
  path        = "pki/root"
  type        = "pki"
  description = "Nexus Platform Root CA"
  default_lease_ttl_seconds = 87600
  max_lease_ttl_seconds     = 315360000
}

resource "vault_pki_secret_backend_root_cert" "root_ca" {
  depends_on = [vault_mount.root_ca]
  backend     = vault_mount.root_ca.path
  type        = "internal"
  common_name = "Nexus Hospitality Platform Root CA"
  ttl         = "315360000"
  key_type    = "rsa"
  key_bits    = 4096
  ou          = "Security Operations"
  organization = "Nexus Hospitality Technologies"
  country     = "US"
  province    = "Delaware"
  locality    = "Wilmington"
}

resource "vault_pki_secret_backend_config_urls" "root_ca" {
  backend = vault_mount.root_ca.path
  issuing_certificates = [
    "https://vault.nexus-platform.internal/v1/pki/root/ca"
  ]
  crl_distribution_points = [
    "https://vault.nexus-platform.internal/v1/pki/root/crl"
  ]
}

# ============================================================
# INTERMEDIATE CA — PLATFORM SERVICES
# ============================================================

resource "vault_mount" "platform_intermediate" {
  path        = "pki/platform"
  type        = "pki"
  description = "Nexus Platform Services Intermediate CA"
  default_lease_ttl_seconds = 2592000
  max_lease_ttl_seconds     = 94608000
}

resource "vault_pki_secret_backend_intermediate_cert_request" "platform_csr" {
  depends_on = [vault_mount.platform_intermediate]
  backend     = vault_mount.platform_intermediate.path
  type        = "internal"
  common_name = "Nexus Platform Services Intermediate CA"
  key_type    = "rsa"
  key_bits    = 4096
}

resource "vault_pki_secret_backend_root_sign_intermediate" "platform_signed" {
  depends_on = [vault_pki_secret_backend_root_cert.root_ca]
  backend     = vault_mount.root_ca.path
  csr         = vault_pki_secret_backend_intermediate_cert_request.platform_csr.csr
  common_name = "Nexus Platform Services Intermediate CA"
  ttl         = "94608000"
  permitted_dns_domains = [
    "nexus-platform.internal",
    "*.nexus-platform.internal",
    "nexus-streaming.com",
    "*.nexus-streaming.com"
  ]
}

resource "vault_pki_secret_backend_intermediate_set_signed" "platform" {
  backend     = vault_mount.platform_intermediate.path
  certificate = vault_pki_secret_backend_root_sign_intermediate.platform_signed.certificate
}

# Role for service-to-service mTLS
resource "vault_pki_secret_backend_role" "service_mtls" {
  backend = vault_mount.platform_intermediate.path
  name    = "service-mtls"
  allowed_domains = [
    "nexus-platform.internal",
    "*.svc.cluster.local",
    "*.nexus-streaming.com"
  ]
  allow_subdomains = true
  allow_glob_domains = true
  key_type     = "ec"
  key_bits     = 256
  signature_algorithm = "ecdsa-with-SHA256"
  ttl          = "7776000"
  max_ttl      = "7776000"
  generate_lease = true
  require_cn = true
  enforce_hostnames = true
  ou = ["Platform Services"]
  organization = ["Nexus Hospitality Technologies"]
  key_usage = ["DigitalSignature", "KeyAgreement"]
  ext_key_usage = ["ServerAuth", "ClientAuth"]
}

# Role for ingress TLS
resource "vault_pki_secret_backend_role" "ingress_tls" {
  backend = vault_mount.platform_intermediate.path
  name    = "ingress-tls"
  allowed_domains = [
    "nexus-platform.com",
    "*.nexus-platform.com",
    "nexus-streaming.com",
    "*.nexus-streaming.com"
  ]
  allow_wildcard_certificates = true
  allow_subdomains = true
  key_type = "rsa"
  key_bits = 2048
  ttl     = "2592000"
  max_ttl = "7776000"
  require_cn = true
  key_usage = ["DigitalSignature", "KeyEncipherment"]
  ext_key_usage = ["ServerAuth"]
}

# ============================================================
# INTERMEDIATE CA — IOT DEVICE IDENTITY
# ============================================================

resource "vault_mount" "iot_intermediate" {
  path        = "pki/iot"
  type        = "pki"
  description = "Nexus IoT Device Identity Intermediate CA"
  default_lease_ttl_seconds = 604800
  max_lease_ttl_seconds     = 31536000
}

resource "vault_pki_secret_backend_intermediate_cert_request" "iot_csr" {
  depends_on = [vault_mount.iot_intermediate]
  backend     = vault_mount.iot_intermediate.path
  type        = "internal"
  common_name = "Nexus IoT Device Identity Intermediate CA"
  key_type    = "ec"
  key_bits    = 384
}

resource "vault_pki_secret_backend_root_sign_intermediate" "iot_signed" {
  backend     = vault_mount.root_ca.path
  csr         = vault_pki_secret_backend_intermediate_cert_request.iot_csr.csr
  common_name = "Nexus IoT Device Identity Intermediate CA"
  ttl         = "31536000"
  permitted_dns_domains = []
  permitted_uri_domains = [
    "nexus-iot.internal"
  ]
}

resource "vault_pki_secret_backend_intermediate_set_signed" "iot" {
  backend     = vault_mount.iot_intermediate.path
  certificate = vault_pki_secret_backend_root_sign_intermediate.iot_signed.certificate
}

resource "vault_pki_secret_backend_role" "iot_device" {
  backend = vault_mount.iot_intermediate.path
  name    = "iot-device"
  allowed_uri_sans = [
    "nexus-iot.internal/device/*"
  ]
  allow_any_name = false
  key_type = "ec"
  key_bits = 256
  ttl     = "2592000"
  max_ttl = "7776000"
  require_cn = true
  key_usage = ["DigitalSignature"]
  ext_key_usage = ["ClientAuth"]
}

# ============================================================
# INTERMEDIATE CA — GUEST/END-USER IDENTITY
# ============================================================

resource "vault_mount" "guest_intermediate" {
  path        = "pki/guest"
  type        = "pki"
  description = "Nexus Guest Identity Intermediate CA"
  default_lease_ttl_seconds = 86400
  max_lease_ttl_seconds     = 604800
}

resource "vault_pki_secret_backend_intermediate_cert_request" "guest_csr" {
  depends_on = [vault_mount.guest_intermediate]
  backend     = vault_mount.guest_intermediate.path
  type        = "internal"
  common_name = "Nexus Guest Identity Intermediate CA"
  key_type    = "ec"
  key_bits    = 256
}

resource "vault_pki_secret_backend_root_sign_intermediate" "guest_signed" {
  backend     = vault_mount.root_ca.path
  csr         = vault_pki_secret_backend_intermediate_cert_request.guest_csr.csr
  common_name = "Nexus Guest Identity Intermediate CA"
  ttl         = "604800"
}

resource "vault_pki_secret_backend_intermediate_set_signed" "guest" {
  backend     = vault_mount.guest_intermediate.path
  certificate = vault_pki_secret_backend_root_sign_intermediate.guest_signed.certificate
}

resource "vault_pki_secret_backend_role" "guest_session" {
  backend = vault_mount.guest_intermediate.path
  name    = "guest-session"
  allowed_uri_sans = [
    "nexus-guest.internal/session/*"
  ]
  key_type = "ec"
  key_bits = 256
  ttl     = "86400"
  max_ttl = "604800"
  require_cn = true
  key_usage = ["DigitalSignature"]
  ext_key_usage = ["ClientAuth"]
}

# ============================================================
# CERTIFICATE ROTATION AUTOMATION
# ============================================================

resource "vault_policy" "cert_rotation" {
  name = "cert-rotation-policy"
  policy = <<EOT
path "pki/platform/cert/ca" {
  capabilities = ["read"]
}
path "pki/platform/issue/service-mtls" {
  capabilities = ["create", "update"]
}
path "pki/platform/issue/ingress-tls" {
  capabilities = ["create", "update"]
}
path "pki/platform/cert/*" {
  capabilities = ["read"]
}
path "pki/platform/revoke" {
  capabilities = ["create", "update"]
}
path "pki/iot/issue/iot-device" {
  capabilities = ["create", "update"]
}
path "pki/iot/revoke" {
  capabilities = ["create", "update"]
}
EOT
}

# Kubernetes auth for cert-manager
resource "vault_auth_backend" "kubernetes" {
  type = "kubernetes"
  path = "kubernetes"
}

resource "vault_kubernetes_auth_backend_config" "k8s" {
  backend                = vault_auth_backend.kubernetes.path
  kubernetes_host        = "https://kubernetes.default.svc"
  token_reviewer_jwt     = file("/var/run/secrets/kubernetes.io/serviceaccount/token")
}

resource "vault_kubernetes_auth_backend_role" "cert_manager" {
  backend                          = vault_auth_backend.kubernetes.path
  role_name                        = "cert-manager"
  bound_service_account_names      = ["cert-manager"]
  bound_service_account_namespaces = ["cert-manager"]
  token_ttl                        = 3600
  token_policies                   = [vault_policy.cert_rotation.name]
}

# ============================================================
# CERTIFICATE MONITORING & ALERTING
# ============================================================

resource "vault_audit" "pki_audit" {
  type = "file"
  options = {
    file_path = "/var/log/vault/audit.log"
  }
}

resource "vault_generic_endpoint" "pki_metrics" {
  depends_on = [vault_mount.platform_intermediate]
  path = "sys/metrics"
  data_json = jsonencode({
    enabled = true
    prometheus_retention_time = "30s"
    disable_hostname = true
  })
}

# ============================================================
# AUTOMATIC CRL DISTRIBUTION
# ============================================================

resource "vault_pki_secret_backend_crl_config" "platform_crl" {
  backend = vault_mount.platform_intermediate.path
  expiry  = "72h"
  disable = false
  auto_rebuild = true
  auto_rebuild_grace_period = "12h"
}

resource "vault_pki_secret_backend_crl_config" "iot_crl" {
  backend = vault_mount.iot_intermediate.path
  expiry  = "24h"
  disable = false
  auto_rebuild = true
  auto_rebuild_grace_period = "4h"
}
