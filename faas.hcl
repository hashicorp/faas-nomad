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
        NOMAD_ADDR   = "${NOMAD_IP_http}:4646"
        CONSUL_ADDR  = "${NOMAD_IP_http}:8500"
        HOST_IP      = "${NOMAD_IP_http}"
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
        functions_provider_url = "http://${NOMAD_IP_http}:8080/"
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
