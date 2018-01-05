job "fabio" {
  datacenters = ["dc1"]

  type = "system"

  constraint {
    attribute = "${attr.cpu.arch}"
    operator  = "="
    value     = "amd64"
  }

  group "fabio" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "fabio" {
      driver = "docker"
      
      env = {
        registry.consul.addr = "${NOMAD_IP_http}:8500"
      }

      config {
        image = "magiconair/fabio:1.5.3-go1.9.2"

        port_map {
          http = 9999
          admin = 9998
        }

      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 256 # 256MB

        network {
          mbits = 10

          port "admin" {
           static = 9998
          }

          port "http" {
            static = 80
          }
        }
      }

      service {
        port = "admin"
        name = "faasd-fabio"
        tags = ["faas"]

        check {
          name     = "alive"
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          path     = "/health"
        }
      }
    }
  }
}
