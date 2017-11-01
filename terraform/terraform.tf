module "open-faas-nomad" {
  source = "nicholasjackson/open-faas-nomad/aws"

  namespace = "openfaas"

  instance_type = "t2.medium"

  ssh_key = "~/.ssh/id_rsa.pub"

  min_servers = "1"
  max_servers = "5"
  min_agents  = "3"
  max_agents  = "5"

  consul_version = "1.0.0"
  nomad_version  = "0.7.0"
}

output "nomad_alb" {
  value = "${module.open-faas-nomad.nomad_alb}"
}

output "faas_alb" {
  value = "${module.open-faas-nomad.faas_alb}"
}

output "grafana_alb" {
  value = "${module.open-faas-nomad.grafana_alb}"
}

output "prometheus_alb" {
  value = "${module.open-faas-nomad.prometheus_alb}"
}
