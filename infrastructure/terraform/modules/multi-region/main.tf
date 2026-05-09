terraform {
  required_version = ">= 1.6.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
  }

  backend "s3" {
    bucket         = "nexus-terraform-state"
    key            = "global/multi-region.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "nexus-terraform-locks"
  }
}

variable "regions" {
  description = "List of deployment regions with configuration"
  type = map(object({
    region_code       = string
    provider          = string
    sovereign         = bool
    compliance_zone   = string
    latency_target_ms = number
    capacity = object({
      min_nodes  = number
      max_nodes  = number
      gpu_nodes  = number
      edge_nodes = number
    })
    networking = object({
      vpc_cidr     = string
      pod_cidr     = string
      service_cidr = string
    })
    disaster_recovery = object({
      paired_region = string
      rpo_minutes   = number
      rto_minutes   = number
    })
  }))

  default = {
    "us-east-1" = {
      region_code       = "use1"
      provider          = "aws"
      sovereign         = false
      compliance_zone   = "pci-dss"
      latency_target_ms = 50
      capacity = {
        min_nodes  = 10
        max_nodes  = 500
        gpu_nodes  = 20
        edge_nodes = 10
      }
      networking = {
        vpc_cidr     = "10.0.0.0/16"
        pod_cidr     = "10.1.0.0/16"
        service_cidr = "10.2.0.0/16"
      }
      disaster_recovery = {
        paired_region = "us-west-2"
        rpo_minutes   = 5
        rto_minutes   = 15
      }
    }
    "us-west-2" = {
      region_code       = "usw2"
      provider          = "aws"
      sovereign         = false
      compliance_zone   = "pci-dss"
      latency_target_ms = 50
      capacity = {
        min_nodes  = 10
        max_nodes  = 300
        gpu_nodes  = 10
        edge_nodes = 8
      }
      networking = {
        vpc_cidr     = "10.10.0.0/16"
        pod_cidr     = "10.11.0.0/16"
        service_cidr = "10.12.0.0/16"
      }
      disaster_recovery = {
        paired_region = "us-east-1"
        rpo_minutes   = 5
        rto_minutes   = 15
      }
    }
    "eu-west-1" = {
      region_code       = "euw1"
      provider          = "aws"
      sovereign         = false
      compliance_zone   = "gdpr"
      latency_target_ms = 50
      capacity = {
        min_nodes  = 10
        max_nodes  = 400
        gpu_nodes  = 15
        edge_nodes = 10
      }
      networking = {
        vpc_cidr     = "10.20.0.0/16"
        pod_cidr     = "10.21.0.0/16"
        service_cidr = "10.22.0.0/16"
      }
      disaster_recovery = {
        paired_region = "eu-central-1"
        rpo_minutes   = 5
        rto_minutes   = 15
      }
    }
    "eu-central-1" = {
      region_code       = "euc1"
      provider          = "aws"
      sovereign         = true
      compliance_zone   = "gdpr-sovereign"
      latency_target_ms = 50
      capacity = {
        min_nodes  = 5
        max_nodes  = 200
        gpu_nodes  = 5
        edge_nodes = 5
      }
      networking = {
        vpc_cidr     = "10.30.0.0/16"
        pod_cidr     = "10.31.0.0/16"
        service_cidr = "10.32.0.0/16"
      }
      disaster_recovery = {
        paired_region = "eu-west-1"
        rpo_minutes   = 5
        rto_minutes   = 15
      }
    }
    "ap-southeast-1" = {
      region_code       = "apse1"
      provider          = "aws"
      sovereign         = false
      compliance_zone   = "gdpr"
      latency_target_ms = 80
      capacity = {
        min_nodes  = 5
        max_nodes  = 300
        gpu_nodes  = 10
        edge_nodes = 8
      }
      networking = {
        vpc_cidr     = "10.40.0.0/16"
        pod_cidr     = "10.41.0.0/16"
        service_cidr = "10.42.0.0/16"
      }
      disaster_recovery = {
        paired_region = "ap-northeast-1"
        rpo_minutes   = 10
        rto_minutes   = 30
      }
    }
    "me-south-1" = {
      region_code       = "mes1"
      provider          = "aws"
      sovereign         = true
      compliance_zone   = "sovereign"
      latency_target_ms = 100
      capacity = {
        min_nodes  = 3
        max_nodes  = 100
        gpu_nodes  = 2
        edge_nodes = 3
      }
      networking = {
        vpc_cidr     = "10.50.0.0/16"
        pod_cidr     = "10.51.0.0/16"
        service_cidr = "10.52.0.0/16"
      }
      disaster_recovery = {
        paired_region = "eu-central-1"
        rpo_minutes   = 15
        rto_minutes   = 60
      }
    }
  }
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "tenant_tiers" {
  description = "Tenant tier configurations"
  type = map(object({
    resource_quota = object({
      cpu_request    = string
      cpu_limit      = string
      memory_request = string
      memory_limit   = string
      gpu_request    = string
      storage        = string
      pods           = number
    })
    network_policy    = string
    isolation_level   = string
  }))
  default = {
    "standard" = {
      resource_quota = {
        cpu_request    = "10"
        cpu_limit      = "50"
        memory_request = "20Gi"
        memory_limit   = "100Gi"
        gpu_request    = "0"
        storage        = "500Gi"
        pods           = 100
      }
      network_policy  = "default-deny-ingress"
      isolation_level = "shared"
    }
    "enterprise" = {
      resource_quota = {
        cpu_request    = "100"
        cpu_limit      = "500"
        memory_request = "200Gi"
        memory_limit   = "1Ti"
        gpu_request    = "4"
        storage        = "5Ti"
        pods           = 1000
      }
      network_policy  = "tenant-isolated"
      isolation_level = "dedicated_namespace"
    }
    "sovereign" = {
      resource_quota = {
        cpu_request    = "500"
        cpu_limit      = "2000"
        memory_request = "1Ti"
        memory_limit   = "4Ti"
        gpu_request    = "20"
        storage        = "20Ti"
        pods           = 5000
      }
      network_policy  = "air-gapped"
      isolation_level = "dedicated_cluster"
    }
  }
}

provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "terraform"
      Platform    = "nexus-hospitality"
    }
  }
}

provider "aws" {
  alias  = "us-west-2"
  region = "us-west-2"
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "terraform"
      Platform    = "nexus-hospitality"
    }
  }
}

provider "aws" {
  alias  = "eu-west-1"
  region = "eu-west-1"
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "terraform"
      Platform    = "nexus-hospitality"
    }
  }
}

provider "aws" {
  alias  = "eu-central-1"
  region = "eu-central-1"
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "terraform"
      Platform    = "nexus-hospitality"
    }
  }
}

provider "aws" {
  alias  = "ap-southeast-1"
  region = "ap-southeast-1"
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "terraform"
      Platform    = "nexus-hospitality"
    }
  }
}

provider "aws" {
  alias  = "me-south-1"
  region = "me-south-1"
  default_tags {
    tags = {
      Environment = var.environment
      ManagedBy   = "terraform"
      Platform    = "nexus-hospitality"
    }
  }
}

module "region_us_east_1" {
  source = "./modules/regional-platform"
  providers = {
    aws = aws.us-east-1
  }
  region_config   = var.regions["us-east-1"]
  environment     = var.environment
  tenant_tiers    = var.tenant_tiers
  global_dns_zone = aws_route53_zone.nexus_global.name
  peer_vpc_ids = [
    module.region_us_west_2.vpc_id,
    module.region_eu_west_1.vpc_id,
  ]
}

module "region_us_west_2" {
  source = "./modules/regional-platform"
  providers = {
    aws = aws.us-west-2
  }
  region_config   = var.regions["us-west-2"]
  environment     = var.environment
  tenant_tiers    = var.tenant_tiers
  global_dns_zone = aws_route53_zone.nexus_global.name
  peer_vpc_ids = [
    module.region_us_east_1.vpc_id,
  ]
}

module "region_eu_west_1" {
  source = "./modules/regional-platform"
  providers = {
    aws = aws.eu-west-1
  }
  region_config   = var.regions["eu-west-1"]
  environment     = var.environment
  tenant_tiers    = var.tenant_tiers
  global_dns_zone = aws_route53_zone.nexus_global.name
  peer_vpc_ids = [
    module.region_us_east_1.vpc_id,
    module.region_eu_central_1.vpc_id,
  ]
}

module "region_eu_central_1" {
  source = "./modules/regional-platform"
  providers = {
    aws = aws.eu-central-1
  }
  region_config   = var.regions["eu-central-1"]
  environment     = var.environment
  tenant_tiers    = var.tenant_tiers
  global_dns_zone = aws_route53_zone.nexus_global.name
  peer_vpc_ids = [
    module.region_eu_west_1.vpc_id,
  ]
}

module "region_ap_southeast_1" {
  source = "./modules/regional-platform"
  providers = {
    aws = aws.ap-southeast-1
  }
  region_config   = var.regions["ap-southeast-1"]
  environment     = var.environment
  tenant_tiers    = var.tenant_tiers
  global_dns_zone = aws_route53_zone.nexus_global.name
  peer_vpc_ids    = []
}

module "region_me_south_1" {
  source = "./modules/regional-platform"
  providers = {
    aws = aws.me-south-1
  }
  region_config   = var.regions["me-south-1"]
  environment     = var.environment
  tenant_tiers    = var.tenant_tiers
  global_dns_zone = aws_route53_zone.nexus_global.name
  peer_vpc_ids = [
    module.region_eu_central_1.vpc_id,
  ]
}

resource "aws_route53_zone" "nexus_global" {
  provider = aws.us-east-1
  name     = var.environment == "production" ? "nexus-platform.global" : "${var.environment}.nexus-platform.global"
}

resource "aws_route53_record" "api_global" {
  provider = aws.us-east-1
  zone_id  = aws_route53_zone.nexus_global.zone_id
  name     = "api"
  type     = "A"
  latency_routing_policy {
    region = "us-east-1"
  }
  set_identifier    = "us-east-1"
  health_check_id   = aws_route53_health_check.us_east_1.id
  alias {
    name                   = module.region_us_east_1.api_endpoint
    zone_id                = module.region_us_east_1.api_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "api_global_usw2" {
  provider = aws.us-east-1
  zone_id  = aws_route53_zone.nexus_global.zone_id
  name     = "api"
  type     = "A"
  latency_routing_policy {
    region = "us-west-2"
  }
  set_identifier    = "us-west-2"
  health_check_id   = aws_route53_health_check.us_west_2.id
  alias {
    name                   = module.region_us_west_2.api_endpoint
    zone_id                = module.region_us_west_2.api_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_record" "api_global_euw1" {
  provider = aws.us-east-1
  zone_id  = aws_route53_zone.nexus_global.zone_id
  name     = "api"
  type     = "A"
  latency_routing_policy {
    region = "eu-west-1"
  }
  set_identifier    = "eu-west-1"
  health_check_id   = aws_route53_health_check.eu_west_1.id
  alias {
    name                   = module.region_eu_west_1.api_endpoint
    zone_id                = module.region_eu_west_1.api_zone_id
    evaluate_target_health = true
  }
}

resource "aws_route53_health_check" "us_east_1" {
  provider          = aws.us-east-1
  fqdn              = module.region_us_east_1.health_check_endpoint
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 30
  regions           = ["us-east-1", "us-west-1", "us-west-2"]
  tags = {
    Name = "nexus-health-use1"
  }
}

resource "aws_route53_health_check" "us_west_2" {
  provider          = aws.us-east-1
  fqdn              = module.region_us_west_2.health_check_endpoint
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 30
  regions           = ["us-west-1", "us-west-2", "us-east-1"]
  tags = {
    Name = "nexus-health-usw2"
  }
}

resource "aws_route53_health_check" "eu_west_1" {
  provider          = aws.us-east-1
  fqdn              = module.region_eu_west_1.health_check_endpoint
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 30
  regions           = ["eu-west-1", "eu-west-2", "eu-central-1"]
  tags = {
    Name = "nexus-health-euw1"
  }
}

resource "aws_cloudwatch_dashboard" "global_operations" {
  provider       = aws.us-east-1
  dashboard_name = "nexus-global-operations-${var.environment}"
  dashboard_body = jsonencode({
    widgets = [
      {
        type   = "metric"
        x      = 0
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "Global Request Latency (p99)"
          region = "us-east-1"
          metrics = [
            for region, config in var.regions : [
              "AWS/ApplicationELB", "TargetResponseTime",
              "LoadBalancer", module["region_${replace(region, "-", "_")}"].alb_arn_suffix,
              { stat = "p99", region = region, label = region }
            ]
          ]
          period = 60
          yAxis = {
            left = {
              min = 0
              max = 500
            }
          }
        }
      },
      {
        type   = "metric"
        x      = 12
        y      = 0
        width  = 12
        height = 6
        properties = {
          title  = "Global Error Rate"
          region = "us-east-1"
          metrics = [
            for region, config in var.regions : [
              "AWS/ApplicationELB", "HTTPCode_Target_5XX_Count",
              "LoadBalancer", module["region_${replace(region, "-", "_")}"].alb_arn_suffix,
              { stat = "Sum", region = region, label = region }
            ]
          ]
          period = 60
        }
      },
      {
        type   = "metric"
        x      = 0
        y      = 6
        width  = 24
        height = 6
        properties = {
          title  = "Active IPTV Streams by Region"
          region = "us-east-1"
          metrics = [
            for region, config in var.regions : [
              "Nexus/IPTV", "ActiveStreams",
              "Region", region,
              { stat = "Average", region = region, label = region }
            ]
          ]
          period = 60
        }
      }
    ]
  })
}

resource "aws_ce_cost_allocation_tag" "tenant_id" {
  provider = aws.us-east-1
  tag_key  = "TenantID"
  status   = "Active"
}

resource "aws_ce_cost_allocation_tag" "service_name" {
  provider = aws.us-east-1
  tag_key  = "ServiceName"
  status   = "Active"
}

resource "aws_budgets_budget" "global_monthly" {
  provider     = aws.us-east-1
  name         = "nexus-global-monthly-${var.environment}"
  budget_type  = "COST"
  limit_amount = var.environment == "production" ? "500000" : "50000"
  limit_unit   = "USD"
  time_period_start = "2024-01-01_00:00"
  time_unit    = "MONTHLY"
  cost_filter {
    name = "TagKeyValue"
    values = [
      "user:Environment$${var.environment}",
      "user:Platform$nexus-hospitality"
    ]
  }
  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 80
    threshold_type             = "PERCENTAGE"
    notification_type          = "ACTUAL"
    subscriber_email_addresses = ["finops@nexus-platform.com"]
  }
  notification {
    comparison_operator        = "GREATER_THAN"
    threshold                  = 100
    threshold_type             = "PERCENTAGE"
    notification_type          = "FORECASTED"
    subscriber_email_addresses = ["cto@nexus-platform.com", "finops@nexus-platform.com"]
  }
}

output "global_dns_zone" {
  description = "Global DNS zone name"
  value       = aws_route53_zone.nexus_global.name
}

output "regional_endpoints" {
  description = "API endpoints by region"
  value = {
    for region, mod in {
      "us-east-1"      = module.region_us_east_1
      "us-west-2"      = module.region_us_west_2
      "eu-west-1"      = module.region_eu_west_1
      "eu-central-1"   = module.region_eu_central_1
      "ap-southeast-1" = module.region_ap_southeast_1
      "me-south-1"     = module.region_me_south_1
    } : region => {
      api_endpoint       = mod.api_endpoint
      streaming_endpoint = mod.streaming_endpoint
      iot_endpoint       = mod.iot_endpoint
      cluster_endpoint   = mod.cluster_endpoint
    }
  }
}

output "disaster_recovery_pairs" {
  description = "DR region pairs"
  value = {
    for region, config in var.regions : region => config.disaster_recovery.paired_region
  }
}
