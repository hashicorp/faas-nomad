job "http-echo" {
  datacenters = ["dc1"]

 update {
    max_parallel     = 1
    health_check     = "checks"
    min_healthy_time = "10s"
    healthy_deadline = "10m"
    auto_revert      = true
    # canary           = 1
    stagger          = "30s"
  } type = "service"

  constraint {
    attribute = "${attr.cpu.arch}"
    operator  = "="
    value     = "amd64"
  }

  group "echo" {
    count = 5

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "echo" {
      driver = "docker"
      
      env = {
        registry.consul.addr = "${NOMAD_IP_http}:8500"
      }

      config {
        image = "hashicorp/http-echo"


        args = [
          "-listen", ":8083", 
          "-text", "hello world v1" 
        ]

        port_map {
          http = 8083
        }

      }

      resources {
        cpu    = 500 # 500 MHz
        memory = 256 # 256MB

        network {
          mbits = 10

          port "http" {
          }
        }
      }

      service {
        port = "http"
        name = "faasd-echo"
        tags = [
          "urlprefix-/",
        ]
        check {
          name     = "alive"
          type     = "http"
          interval = "10s"
          timeout  = "2s"
          path     = "/"
        }
      }
    }
  }
}
