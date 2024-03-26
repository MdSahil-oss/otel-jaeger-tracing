variable "vpc_id" {
  type    = string
  default = "vpc-05d72870c9ff5a6dc"
}

variable "cluster_name" {
  type    = string
  default = "mdsahiloss-cluster"
}

locals {
  region = "us-east-1"
  zoneA  = "us-east-1a"
  zoneB  = "us-east-1b"
}
