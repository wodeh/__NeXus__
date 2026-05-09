# terraform/modules/eks/main.tf
# Production EKS cluster for Nexus Hospitality Cloud

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  cluster_name = "nhc-${var.environment}-${var.region}"
  tags = merge(var.common_tags, {
    Environment = var.environment
    Region      = var.region
    ManagedBy   = "terraform"
  })
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${local.cluster_name}-vpc"
  cidr = var.vpc_cidr

  azs             = var.availability_zones
  private_subnets = var.private_subnet_cidrs
  public_subnets  = var.public_subnet_cidrs
  intra_subnets   = var.intra_subnet_cidrs

  enable_nat_gateway     = true
  single_nat_gateway     = var.environment != "production"
  enable_dns_hostnames   = true
  enable_dns_support     = true

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
    "karpenter.sh/discovery"          = local.cluster_name
  }

  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  tags = local.tags
}

module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = local.cluster_name
  cluster_version = "1.29"

  cluster_endpoint_public_access  = false
  cluster_endpoint_private_access = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  cluster_addons = {
    coredns = {
      most_recent = true
      configuration_values = jsonencode({ computeType = "Fargate" })
    }
    kube-proxy = { most_recent = true }
    vpc-cni = {
      most_recent = true
      configuration_values = jsonencode({
        env = {
          ENABLE_PREFIX_DELEGATION = "true"
          WARM_PREFIX_TARGET       = "1"
        }
      })
    }
    aws-ebs-csi-driver = { most_recent = true }
    aws-efs-csi-driver = { most_recent = true }
    amazon-cloudwatch-observability = { most_recent = true }
  }

  eks_managed_node_groups = {
    general = {
      desired_size = var.general_node_count
      min_size     = max(2, var.general_node_count / 2)
      max_size     = var.general_node_max
      instance_types = var.general_instance_types
      capacity_type  = "ON_DEMAND"
      labels = { workload = "general", node-type = "compute" }
      block_device_mappings = {
        xvda = {
          device_name = "/dev/xvda"
          ebs = {
            volume_size           = 100
            volume_type           = "gp3"
            iops                  = 3000
            throughput            = 125
            encrypted             = true
            kms_key_id            = aws_kms_key.eks.arn
            delete_on_termination = true
          }
        }
      }
    }
    compute = {
      desired_size = var.compute_node_count
      min_size     = max(1, var.compute_node_count / 2)
      max_size     = var.compute_node_max
      instance_types = var.compute_instance_types
      capacity_type  = "ON_DEMAND"
      labels = { workload = "compute-intensive", node-type = "compute" }
    }
    gpu = {
      desired_size = var.gpu_node_count
      min_size     = 0
      max_size     = var.gpu_node_max
      instance_types = var.gpu_instance_types
      capacity_type  = "ON_DEMAND"
      labels = { workload = "ai-ml", "nvidia.com/gpu" = "true" }
      taints = [{
        key    = "nvidia.com/gpu"
        value  = "true"
        effect = "NO_SCHEDULE"
      }]
    }
    spot = {
      desired_size = var.spot_node_count
      min_size     = 0
      max_size     = var.spot_node_max
      instance_types = var.spot_instance_types
      capacity_type  = "SPOT"
      labels = { workload = "batch", node-type = "spot" }
      taints = [{
        key    = "spot"
        value  = "true"
        effect = "NO_SCHEDULE"
      }]
    }
  }

  fargate_profiles = {
    observability = {
      name = "observability"
      selectors = [
        { namespace = "nexus-observability" },
        { namespace = "nexus-jobs" }
      ]
      subnet_ids = module.vpc.private_subnets
    }
  }

  cluster_security_group_additional_rules = {
    ingress_nodes_ephemeral_ports_tcp = {
      description                = "Nodes on ephemeral ports"
      protocol                   = "tcp"
      from_port                  = 1025
      to_port                    = 65535
      type                       = "ingress"
      source_node_security_group = true
    }
  }

  tags = local.tags
}

resource "aws_kms_key" "eks" {
  description             = "EKS Secret Encryption Key"
  deletion_window_in_days = 7
  enable_key_rotation     = true
  multi_region            = var.environment == "production"
  tags = local.tags
}

resource "aws_kms_alias" "eks" {
  name          = "alias/nhc-eks-${var.environment}"
  target_key_id = aws_kms_key.eks.key_id
}

module "irsa_iptv_session_manager" {
  source  = "terraform-aws-modules/iam/aws//modules/iam-role-for-service-accounts-eks"
  version = "~> 5.0"
  role_name = "nhc-iptv-session-manager-${var.environment}"
  oidc_providers = {
    main = {
      provider_arn               = module.eks.oidc_provider_arn
      namespace_service_accounts = ["nexus-iptv:iptv-session-manager"]
    }
  }
  role_policy_arns = {
    dynamodb        = aws_iam_policy.iptv_dynamodb.arn
    s3              = aws_iam_policy.iptv_s3.arn
    secrets_manager = aws_iam_policy.iptv_secrets.arn
    cloudwatch      = aws_iam_policy.iptv_cloudwatch.arn
  }
}

resource "aws_iam_policy" "iptv_dynamodb" {
  name = "nhc-iptv-dynamodb-${var.environment}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["dynamodb:GetItem", "dynamodb:PutItem", "dynamodb:UpdateItem", "dynamodb:DeleteItem", "dynamodb:Query", "dynamodb:Scan"]
      Resource = "arn:aws:dynamodb:*:*:table/nhc-iptv-*"
    }]
  })
}

resource "aws_iam_policy" "iptv_s3" {
  name = "nhc-iptv-s3-${var.environment}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject", "s3:ListBucket"]
      Resource = ["arn:aws:s3:::nhc-iptv-*", "arn:aws:s3:::nhc-iptv-*/*"]
    }]
  })
}

resource "aws_iam_policy" "iptv_secrets" {
  name = "nhc-iptv-secrets-${var.environment}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["secretsmanager:GetSecretValue", "secretsmanager:DescribeSecret"]
      Resource = "arn:aws:secretsmanager:*:*:secret:nhc/iptv/*"
    }]
  })
}

resource "aws_iam_policy" "iptv_cloudwatch" {
  name = "nhc-iptv-cloudwatch-${var.environment}"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents", "cloudwatch:PutMetricData"]
      Resource = "*"
    }]
  })
}

output "cluster_endpoint" { value = module.eks.cluster_endpoint }
output "cluster_name" { value = module.eks.cluster_name }
output "oidc_provider_arn" { value = module.eks.oidc_provider_arn }
output "vpc_id" { value = module.vpc.vpc_id }
