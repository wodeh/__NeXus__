variable "environment" {
  description = "Environment name"
  type        = string
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "vpc_cidr" {
  description = "VPC CIDR block"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "private_subnet_cidrs" {
  type    = list(string)
  default = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "public_subnet_cidrs" {
  type    = list(string)
  default = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
}

variable "intra_subnet_cidrs" {
  type    = list(string)
  default = ["10.0.201.0/24", "10.0.202.0/24", "10.0.203.0/24"]
}

variable "general_node_count" { type = number; default = 5 }
variable "general_node_max" { type = number; default = 50 }
variable "general_instance_types" { type = list(string); default = ["m6i.2xlarge"] }
variable "compute_node_count" { type = number; default = 3 }
variable "compute_node_max" { type = number; default = 30 }
variable "compute_instance_types" { type = list(string); default = ["c6i.4xlarge"] }
variable "gpu_node_count" { type = number; default = 2 }
variable "gpu_node_max" { type = number; default = 20 }
variable "gpu_instance_types" { type = list(string); default = ["p4d.24xlarge"] }
variable "spot_node_count" { type = number; default = 5 }
variable "spot_node_max" { type = number; default = 100 }
variable "spot_instance_types" { type = list(string); default = ["m6i.xlarge", "m5.xlarge", "m5a.xlarge"] }

variable "common_tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "nexus-hospitality-cloud"
    Owner       = "platform-team"
    CostCenter  = "engineering"
  }
}
