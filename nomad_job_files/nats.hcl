job "natsio" {
  datacenters = ["dc1"]

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
        image = "nats-streaming:0.7.0-linux"

        args = [
          "-store", "file", "-dir", "/tmp/nats",
          "-m", "8222",
          "--cluster_id", "faas-cluster"
        ]

        port_map {
          client = 4222,
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
