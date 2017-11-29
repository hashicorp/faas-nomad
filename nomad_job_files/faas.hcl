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

      env {
        NOMAD_REGION = "${NOMAD_REGION}"
        NOMAD_ADDR   = "${NOMAD_IP_http}:4646"
        CONSUL_ADDR  = "${NOMAD_IP_http}:8500"
        STATSD_ADDR  = "${NOMAD_ADDR_statsd_statsd}"
      }

      config {
        image = "quay.io/nicholasjackson/faas-nomad:0.2.12"

        port_map {
          http = 8080
        }
      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 256 # 256MB

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

      env {
        functions_provider_url = "http://${NOMAD_IP_http}:8081/"
      }

      config {
        image = "functions/gateway:0.6.12"

        port_map {
          http = 8080
        }
      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 256 # 256MB

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
        memory = 128 # 128MB

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
}
