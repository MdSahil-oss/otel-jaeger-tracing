provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      args        = ["eks", "get-token", "--cluster-name", module.eks.cluster_name]
      command     = "aws"
    }
  }
}

resource "helm_release" "gitlab-agent" {
  name             = "gitlab-agent"
  repository       = "https://charts.gitlab.io/"
  chart            = "gitlab-agent"
  create_namespace = true
  namespace        = "gitlab-agent-azul-project"
  set {
    name  = "image.tag"
    value = "v16.10.1"
  }
  set {
    name  = "config.kasAddress"
    value = "wss://kas.gitlab.com"
  }
  set_sensitive {
    name  = "config.token"
    value = file("../secrets/agent-token.txt")
  }
}
