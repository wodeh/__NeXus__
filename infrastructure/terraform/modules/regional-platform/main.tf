# infrastructure/terraform/modules/regional-platform/main.tf
# ============================================================
# REGIONAL PLATFORM MODULE — Per-region EKS + VPC + Services
# ============================================================

terraform {
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
  }
}

variable "region_config" {
  description = "Region-specific configuration"
  type = object({
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
  })
}

variable "environment" {
  type = string
}

variable "tenant_tiers" {
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
}

variable "global_dns_zone" {
  type = string
}

variable "peer_vpc_ids" {
  type    = list(string)
  default = []
}

# Local values
locals {
  region_code = var.region_config.region_code
  vpc_cidr    = var.region_config.networking.vpc_cidr

  common_tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
    Platform    = "nexus-hospitality"
    Region      = local.region_code
    Compliance  = var.region_config.compliance_zone
  }
}

# VPC
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "nexus-${local.region_code}-${var.environment}"
  cidr = local.vpc_cidr

  azs             = data.aws_availability_zones.available.names
  private_subnets = [for i, az in data.aws_availability_zones.available.names : cidrsubnet(local.vpc_cidr, 4, i)]
  public_subnets  = [for i, az in data.aws_availability_zones.available.names : cidrsubnet(local.vpc_cidr, 4, i + 8)]
  intra_subnets   = [for i, az in data.aws_availability_zones.available.names : cidrsubnet(local.vpc_cidr, 4, i + 12)]

  enable_nat_gateway     = true
  single_nat_gateway     = var.environment != "production"
  enable_dns_hostnames   = true
  enable_dns_support     = true
  enable_ipv6            = false

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
    "karpenter.sh/discovery"          = "nexus-${local.region_code}"
  }

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  tags = local.common_tags
}

data "aws_availability_zones" "available" {
  state = "available"
}

# VPC Peering
resource "aws_vpc_peering_connection" "peer" {
  for_each = toset(var.peer_vpc_ids)

  vpc_id        = module.vpc.vpc_id
  peer_vpc_id   = each.value
  auto_accept   = true

  tags = merge(local.common_tags, {
    Name = "nexus-peer-${local.region_code}"
  })
}

# EKS Cluster
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "nexus-${local.region_code}-${var.environment}"
  cluster_version = "1.28"

  cluster_endpoint_public_access  = true
  cluster_endpoint_private_access = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_addons = {
    coredns = {
      most_recent = true
    }
    kube-proxy = {
      most_recent = true
    }
    vpc-cni = {
      most_recent = true
    }
    aws-ebs-csi-driver = {
      most_recent = true
    }
  }

  # Managed Node Groups
  eks_managed_node_groups = {
    general = {
      desired_size = var.region_config.capacity.min_nodes
      min_size     = var.region_config.capacity.min_nodes
      max_size     = var.region_config.capacity.max_nodes

      instance_types = ["m6i.2xlarge", "m6a.2xlarge"]
      capacity_type  = "ON_DEMAND"

      labels = {
        workload-type = "general"
      }

      taints = []

      update_config = {
        max_unavailable_percentage = 25
      }

      tags = local.common_tags
    }

    gpu = {
      desired_size = var.region_config.capacity.gpu_nodes
      min_size     = 0
      max_size     = var.region_config.capacity.gpu_nodes * 2

      instance_types = ["g5.2xlarge", "g5.4xlarge"]
      capacity_type  = "ON_DEMAND"
      ami_type       = "AL2_x86_64_GPU"

      labels = {
        workload-type = "gpu"
        nvidia.com/gpu.present = "true"
      }

      taints = [{
        key    = "nvidia.com/gpu"
        value  = "true"
        effect = "NO_SCHEDULE"
      }]

      tags = local.common_tags
    }

    edge = {
      desired_size = var.region_config.capacity.edge_nodes
      min_size     = 1
      max_size     = var.region_config.capacity.edge_nodes * 2

      instance_types = ["m6i.xlarge"]
      capacity_type  = "SPOT"

      labels = {
        workload-type = "edge"
      }

      tags = local.common_tags
    }
  }

  # Fargate profiles for serverless workloads
  fargate_profiles = {
    default = {
      name = "default"
      selectors = [
        { namespace = "kube-system" },
        { namespace = "nexus-*" }
      ]
    }
  }

  tags = local.common_tags
}

# Karpenter for autoscaling
resource "helm_release" "karpenter" {
  namespace        = "karpenter"
  create_namespace = true

  name       = "karpenter"
  repository = "oci://public.ecr.aws/karpenter"
  chart      = "karpenter"
  version    = "v0.32.0"

  set {
    name  = "settings.clusterName"
    value = module.eks.cluster_name
  }

  set {
    name  = "settings.clusterEndpoint"
    value = module.eks.cluster_endpoint
  }

  set {
    name  = "serviceAccount.annotations.eks\.amazonaws\.com/role-arn"
    value = module.eks.oidc_provider_arn
  }

  depends_on = [module.eks]
}

# Application Load Balancer
module "alb" {
  source  = "terraform-aws-modules/alb/aws"
  version = "~> 9.0"

  name = "nexus-${local.region_code}"

  load_balancer_type = "application"

  vpc_id  = module.vpc.vpc_id
  subnets = module.vpc.public_subnets

  security_groups = [aws_security_group.alb.id]

  listeners = {
    https = {
      port            = 443
      protocol        = "HTTPS"
      certificate_arn = aws_acm_certificate.nexus.arn

      fixed_response = {
        content_type = "text/plain"
        message_body = "OK"
        status_code  = "200"
      }

      rules = {
        api = {
          actions = [{
            type             = "forward"
            target_group_key = "api"
          }]
          conditions = [{
            path_pattern = {
              values = ["/api/*"]
            }
          }]
        }
        streaming = {
          actions = [{
            type             = "forward"
            target_group_key = "streaming"
          }]
          conditions = [{
            path_pattern = {
              values = ["/hls/*", "/dash/*", "/whip", "/whep"]
            }
          }]
        }
      }
    }
  }

  target_groups = {
    api = {
      name_prefix          = "api-"
      protocol             = "HTTP"
      port                 = 8080
      target_type          = "ip"
      deregistration_delay = 30

      health_check = {
        enabled             = true
        healthy_threshold   = 2
        interval            = 15
        matcher             = "200"
        path                = "/health"
        port                = "traffic-port"
        protocol            = "HTTP"
        timeout             = 5
        unhealthy_threshold = 2
      }

      create_attachment = false
    }

    streaming = {
      name_prefix          = "str-"
      protocol             = "HTTP"
      port                 = 8080
      target_type          = "ip"
      deregistration_delay = 10

      health_check = {
        enabled             = true
        healthy_threshold   = 2
        interval            = 10
        matcher             = "200"
        path                = "/health"
        port                = "traffic-port"
        protocol            = "HTTP"
        timeout             = 3
        unhealthy_threshold = 2
      }

      create_attachment = false
    }
  }

  tags = local.common_tags
}

# Security Group for ALB
resource "aws_security_group" "alb" {
  name_prefix = "nexus-alb-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

# ACM Certificate
resource "aws_acm_certificate" "nexus" {
  domain_name               = "*.${var.global_dns_zone}"
  subject_alternative_names = [var.global_dns_zone]
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }

  tags = local.common_tags
}

# RDS PostgreSQL (managed)
module "rds" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 6.0"

  identifier = "nexus-${local.region_code}-${var.environment}"

  engine               = "postgres"
  engine_version       = "16"
  family               = "postgres16"
  major_engine_version = "16"
  instance_class       = "db.r6g.2xlarge"

  allocated_storage     = 500
  max_allocated_storage = 2000

  db_name  = "nexus_platform"
  username = "nexus_admin"
  port     = 5432

  multi_az               = var.environment == "production"
  db_subnet_group_name   = module.vpc.database_subnet_group_name
  vpc_security_group_ids = [aws_security_group.rds.id]

  backup_retention_period = 30
  backup_window          = "03:00-04:00"
  maintenance_window     = "Mon:04:00-Mon:05:00"

  deletion_protection = var.environment == "production"

  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn

  tags = local.common_tags
}

resource "aws_security_group" "rds" {
  name_prefix = "nexus-rds-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [module.eks.cluster_security_group_id]
  }

  tags = local.common_tags
}

resource "aws_kms_key" "rds" {
  description             = "KMS key for RDS encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = local.common_tags
}

# ElastiCache Redis
module "elasticache" {
  source  = "terraform-aws-modules/elasticache/aws"
  version = "~> 1.0"

  cluster_id               = "nexus-${local.region_code}"
  description              = "Nexus Redis cluster"
  node_type                = "cache.r6g.xlarge"
  num_cache_nodes          = var.environment == "production" ? 3 : 1
  engine_version           = "7.0"
  port                     = 6379
  apply_immediately        = true
  snapshot_retention_limit = 7

  subnet_group_name  = module.vpc.elasticache_subnet_group_name
  security_group_ids = [aws_security_group.elasticache.id]

  tags = local.common_tags
}

resource "aws_security_group" "elasticache" {
  name_prefix = "nexus-cache-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [module.eks.cluster_security_group_id]
  }

  tags = local.common_tags
}

# MSK Kafka
resource "aws_msk_cluster" "kafka" {
  cluster_name           = "nexus-${local.region_code}"
  kafka_version          = "3.5.1"
  number_of_broker_nodes = 3

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    client_subnets  = module.vpc.private_subnets
    security_groups = [aws_security_group.msk.id]

    storage_info {
      ebs_storage_info {
        volume_size = 1000
      }
    }
  }

  encryption_info {
    encryption_at_rest_kms_key_arn = aws_kms_key.msk.arn
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  open_monitoring {
    prometheus {
      jmx_exporter {
        enabled_in_broker = true
      }
      node_exporter {
        enabled_in_broker = true
      }
    }
  }

  logging_info {
    broker_logs {
      cloudwatch_logs {
        enabled   = true
        log_group = aws_cloudwatch_log_group.msk.name
      }
    }
  }

  tags = local.common_tags
}

resource "aws_security_group" "msk" {
  name_prefix = "nexus-msk-"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 9094
    to_port         = 9094
    protocol        = "tcp"
    security_groups = [module.eks.cluster_security_group_id]
  }

  tags = local.common_tags
}

resource "aws_kms_key" "msk" {
  description             = "KMS key for MSK encryption"
  deletion_window_in_days = 30
  enable_key_rotation     = true

  tags = local.common_tags
}

resource "aws_cloudwatch_log_group" "msk" {
  name              = "/aws/msk/nexus-${local.region_code}"
  retention_in_days = 30

  tags = local.common_tags
}

# S3 Buckets
resource "aws_s3_bucket" "artifacts" {
  bucket = "nexus-artifacts-${local.region_code}-${var.environment}"

  tags = local.common_tags
}

resource "aws_s3_bucket" "streams" {
  bucket = "nexus-streams-${local.region_code}-${var.environment}"

  tags = local.common_tags
}

resource "aws_s3_bucket" "mlflow" {
  bucket = "nexus-mlflow-${local.region_code}-${var.environment}"

  tags = local.common_tags
}

resource "aws_s3_bucket_versioning" "artifacts" {
  bucket = aws_s3_bucket.artifacts.id
  versioning_configuration {
    status = "Enabled"
  }
}

# Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
}

output "api_endpoint" {
  description = "API ALB endpoint"
  value       = module.alb.dns_name
}

output "streaming_endpoint" {
  description = "Streaming endpoint"
  value       = module.alb.dns_name
}

output "iot_endpoint" {
  description = "IoT MQTT endpoint"
  value       = "mqtts://iot.${var.global_dns_zone}:8883"
}

output "health_check_endpoint" {
  description = "Health check endpoint"
  value       = "health.${var.global_dns_zone}"
}

output "api_zone_id" {
  description = "ALB hosted zone ID"
  value       = module.alb.zone_id
}

output "alb_arn_suffix" {
  description = "ALB ARN suffix for CloudWatch"
  value       = module.alb.arn_suffix
}

output "rds_endpoint" {
  description = "RDS endpoint"
  value       = module.rds.db_instance_endpoint
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = module.elasticache.cluster_endpoint
}

output "kafka_brokers" {
  description = "Kafka broker list"
  value       = aws_msk_cluster.kafka.bootstrap_brokers_tls
}
