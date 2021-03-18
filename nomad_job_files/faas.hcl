job "faas-nomadd" {
  datacenters = ["dc1"]

  type = "system"

  constraint {
    attribute = "${attr.cpu.arch}"
    operator  = "="
    value     = "amd64"
  }

  group "faas-nomadd" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "nomadd" {
      driver = "docker"

      config {
        image = "quay.io/nicholasjackson/faas-nomad:v0.4.3-rc2"

        args = [
          "-nomad_region", "${NOMAD_REGION}",
          "-nomad_addr", "${NOMAD_IP_http}:4646",
          "-consul_addr", "${NOMAD_IP_http}:8500",
          "-statsd_addr", "${NOMAD_ADDR_statsd_statsd}",
          "-node_addr", "${NOMAD_IP_http}",
          "-basic_auth_secret_path", "/secrets",
          "-enable_basic_auth=false",
          "-enable_nomad_tls=true",
          "-nomad_tls_skip_verify"
        ]

        port_map {
          http = 8080
        }
      }
      // basic auth from vault example
      // update -enable_basic_auth=true
      // uncomment below if you have a Vault instance connected to Nomad
//       template {
//         destination   = "secrets/basic-auth-user"
//         data = <<EOH
// {{ with secret "secret/openfaas/auth/credentials" }}{{ .Data.username }}{{ end }}
// EOH
//       }
//       template {
//         destination   = "secrets/basic-auth-password"
//         data = <<EOH
// {{ with secret "secret/openfaas/auth/credentials" }}{{ .Data.password }}{{ end }}
// EOH
//       }

      resources {
        cpu    = 500 # 500 MHz
        memory = 128 # 128MB

        network {
          mbits = 10

          port "http" {
            static = 8081
          }
        }
      }

      service {
        port = "http"
        name = "faasd-nomad"
        tags = ["faas"]
      }
    }

    task "gateway" {
      driver = "docker"
      template {
        env = true
        destination   = "secrets/gateway.env"

        data = <<EOH
functions_provider_url="http://{{ env "NOMAD_IP_http" }}:8081/"
{{ range service "prometheus" }}
faas_prometheus_host="{{ .Address }}"
faas_prometheus_port="{{ .Port }}"{{ end }}
{{ range service "nats" }}
faas_nats_address="{{ .Address }}"
faas_nats_port={{ .Port }}{{ end }}
EOH
      }

      config {
        image = "openfaas/gateway:0.9.14"

        port_map {
          http = 8080
        }
      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 128 # 128MB

        network {
          mbits = 10

          port "http" {
            static = 8080
          }
        }
      }

      service {
        port = "http"
        name = "gateway"
        tags = ["faas"]
      }
    }

    task "statsd" {
      driver = "docker"

      config {
        image = "prom/statsd-exporter:v0.4.0"

        args = [
          "-log.level=debug",
        ]
      }

      resources {
        cpu    = 100 # 100 MHz
        memory = 36 # 36MB

        network {
          mbits = 1

          port "http" {
            static = 9102
          }

          port "statsd" {
            static = 9125
          }
        }
      }

      service {
        port = "http"
        name = "statsd"
        tags = ["faas"]

        check {
          type     = "http"
          port     = "http"
          interval = "10s"
          timeout  = "2s"
          path     = "/"
        }
      }
    }
  }

  group "faas-nats" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "nats" {
      driver = "docker"
      
      config {
        image = "nats-streaming:0.11.2-linux"

        args = [
          "-store", "file", "-dir", "/tmp/nats",
          "-m", "8222",
          "-cid","faas-cluster",
        ]

        port_map {
          client = 4222
          monitoring = 8222
          routing = 6222
        }
      }

      resources {
        cpu    = 400 # 100 MHz
        memory = 128 # 128MB

        network {
          mbits = 1

          port "client" {
            static = 4222
          }

          port "monitoring" {
            static = 8222
          }

          port "routing" {
            static = 6222
          }
        }
      }

      service {
        port = "client"
        name = "nats"
        tags = ["faas"]

        check {
           type     = "http"
           port     = "monitoring"
           path     = "/connz"
           interval = "5s"
           timeout  = "2s"
        }
      }
    }
  }
}
