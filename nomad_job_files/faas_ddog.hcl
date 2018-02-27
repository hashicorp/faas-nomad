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
        logger_format = "json"
        logger_output = "/logs/nomadd.log"
      }

      config {
        image = "quay.io/nicholasjackson/faas-nomad:v0.2.22"

        port_map {
          http = 8080
        }
        
        volumes = [
          # Use relative paths to rebind paths already in the allocation dir
          "../logs/:/logs"
        ]
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
        statsd_server          = "${NOMAD_IP_http}:8125"
        logger_format          = "JSON"
        logger_output          = "/logs/gateway.log"
      }

      config {
        image = "nicholasjackson/gateway:latest-dev"

        port_map {
          http = 8080
        }

        volumes = [
          # Use relative paths to rebind paths already in the allocation dir
          "../logs/:/logs"
        ]
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

      env {
       DD_API_KEY = "DD_API_KEY"
     }

     template {
        destination   = "local/gateway.yaml"
        change_mode   = "signal"
        change_signal = "SIGINT"
        data          = <<EOH
init_config:

instances:
    [{}]
    
#Log section
logs:

    # - type : file (mandatory) type of log input source (tcp / udp / file)
    #   port / path : (mandatory) Set port if type is tcp or udp. Set path if type is file
    #   service : (mandatory) name of the service owning the log
    #   source : (mandatory) attribute that defines which integration is sending the logs
    #   sourcecategory : (optional) Multiple value attribute. Can be used to refine the source attribtue
    #   tags: (optional) add tags to each logs collected

  - type: file
    path: /logs/gateway.log
    service: openfaas-gateway
    source: go
EOH
  }
     
  template {
        destination   = "local/nomadd.yaml"
        change_mode   = "signal"
        change_signal = "SIGINT"
        data          = <<EOH
init_config:

instances:
    [{}]
    
#Log section
logs:

    # - type : file (mandatory) type of log input source (tcp / udp / file)
    #   port / path : (mandatory) Set port if type is tcp or udp. Set path if type is file
    #   service : (mandatory) name of the service owning the log
    #   source : (mandatory) attribute that defines which integration is sending the logs
    #   sourcecategory : (optional) Multiple value attribute. Can be used to refine the source attribtue
    #   tags: (optional) add tags to each logs collected

  - type: file
    path: /logs/nomadd.log
    service: openfaas-nomadd
    source: go
EOH
  }
     
  template {
        destination   = "local/datadog.yaml"
        change_mode   = "signal"
        change_signal = "SIGINT"
        data          = <<EOH
log_enabled: true
EOH
  }

      config {
        image = "datadog/agent:6.0.0-beta.6"

        volumes = [
          "../logs:/logs",
          "local/datadog.yaml:/etc/datadog-agent/datadog.yaml:ro",
          "local/gateway.yaml:/conf.d/go.d/gateway.yaml:ro",
          "local/nomadd.yaml:/conf.d/go.d/nomadd.yaml:ro",
        ]

        args = [
          /*
          "-v", "/var/run/docker.sock:/var/run/docker.sock:ro",
          "-v", "-v /proc/:/host/proc/:ro",
          "-v", "/sys/fs/cgroup/:/host/sys/fs/cgroup:ro",
          */
        ]
      }

      resources {
        cpu    = 100 # 100 MHz
        memory = 128 # 128MB

        network {
          mbits = 1

          port "statsd" {
            static = 8125
          }
        }
      }

      service {
        port = "statsd"
        name = "statsd"
        tags = ["faas"]
      }
    }
  }
}
