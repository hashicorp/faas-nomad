job "faas-nomadd" {
  datacenters = ["dc1"]

  type = "system"

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
        NOMAD_ADDR   = "192.168.65.1:4646"
        CONSUL_ADDR  = "192.168.65.1:8500"
        HOST_IP      = "192.168.65.1"
      }

      config {
        image = "quay.io/nicholasjackson/faas-nomad:0.1"
        
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

    }

    task "gateway" {
      driver = "docker"

      env {
        functions_provider_url = "http://192.168.65.1:8080/"
      }

      config {
        image = "functions/gateway:0.6.1"
        
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

    }
  }
}
