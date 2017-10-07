job "faas-monitoring" {
  datacenters = ["dc1"]

  type = "service"

  group "faas-monitoring" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "alertmanager" {
      driver = "docker"

      template {
        data = <<EOF
global:
  # The smarthost and SMTP sender used for mail notifications.
  smtp_smarthost: 'localhost:25'
  smtp_from: 'alertmanager@example.org'
  smtp_auth_username: 'alertmanager'
  smtp_auth_password: 'password'
  # The auth token for Hipchat.
  hipchat_auth_token: '1234556789'
  # Alternative host for Hipchat.
  hipchat_url: 'https://hipchat.foobar.org/'

# The directory from which notification templates are read.
templates:
- '/etc/alertmanager/template/*.tmpl'

# The root route on which each incoming alert enters.
route:
  # The labels by which incoming alerts are grouped together. For example,
  # multiple alerts coming in for cluster=A and alertname=LatencyHigh would
  # be batched into a single group.
  group_by: ['alertname', 'cluster', 'service']

  # When a new group of alerts is created by an incoming alert, wait at
  # least 'group_wait' to send the initial notification.
  # This way ensures that you get multiple alerts for the same group that start
  # firing shortly after another are batched together on the first
  # notification.
  group_wait: 5s

  # When the first notification was sent, wait 'group_interval' to send a batch
  # of new alerts that started firing for that group.
  group_interval: 10s

  # If an alert has successfully been sent, wait 'repeat_interval' to
  # resend them.
  repeat_interval: 30s

  # A default receiver
  receiver: scale-up

  # All the above attributes are inherited by all child routes and can
  # overwritten on each.

  # The child route trees.
  routes:
  - match:
      service: gateway
      receiver: scale-up
      severity: major

# Inhibition rules allow to mute a set of alerts given that another alert is
# firing.
# We use this to mute any warning-level notifications if the same alert is
# already critical.
inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  # Apply inhibition if the alertname is the same.
  equal: ['alertname', 'cluster', 'service']

receivers:
- name: 'scale-up'
  webhook_configs:
    - url: http://gateway.service.dc1.consul:8080/system/alert
      send_resolved: true
EOF

        destination   = "etc/alertmanager/alertmanager.yml"
        change_mode   = "noop"
        change_signal = "SIGINT"
      }

      config {
        image = "prom/alertmanager:v0.9.1"

        port_map {
          http = 9003
        }

        dns_servers = ["${NOMAD_IP_http}", "8.8.8.8", "8.8.8.4"]

        args = [
          "-config.file=/etc/alertmanager/alertmanager.yml",
          "-storage.path=/alertmanager",
        ]

        volumes = [
          "etc/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml",
        ]
      }

      resources {
        cpu    = 100 # 100 MHz
        memory = 128 # 128MB

        network {
          mbits = 10

          port "http" {}
        }
      }

      service {
        port = "http"
        name = "alertmanager"
        tags = ["faas"]
      }
    }

    task "prometheus" {
      driver = "docker"

      template {
        data = <<EOH
global:
  scrape_interval:     15s 
  evaluation_interval: 15s
  scrape_timeout: 10s

  # Attach these labels to any time series or alerts when communicating with
  # external systems (federation, remote storage, Alertmanager).
  external_labels:
    monitor: 'faas-monitor'

# Load rules once and periodically evaluate them according to the global 'evaluation_interval'.
rule_files:
  - 'alert.rules'

# A scrape configuration containing exactly one endpoint to scrape:
# Here it's Prometheus itself.
scrape_configs:
  # The job name is added as a label `job=<job_name>` to any timeseries scraped from this config.
  - job_name: 'prometheus'
    # Override the global default and scrape targets from this job every 5 seconds.
    scrape_interval: 5s

    # metrics_path defaults to '/metrics'
    # scheme defaults to 'http'.
    static_configs:
      - targets: ['localhost:9090']

  - job_name: "gateway"
    scrape_interval: 5s
    dns_sd_configs:
      - names: ["gateway.service.dc1.consul"]
        type: SRV
        refresh_interval: 5s
  
  - job_name: "statsd"
    scrape_interval: 5s
    dns_sd_configs:
      - names: ["statsd.service.dc1.consul"]
        type: SRV
        refresh_interval: 5s
EOH

        destination   = "/etc/prometheus/prometheus.yml"
        change_mode   = "noop"
        change_signal = "SIGINT"
      }

      template {
        data = <<EOH
ALERT service_down
  IF up == 0
ALERT APIHighInvocationRate
  IF sum ( rate(gateway_function_invocation_total{code="200"}[10s]) ) by (function_name) > 5
  FOR 5s
  LABELS {
    service = "gateway",
    severity = "major",
    value = "{{ "{{" }}$value{{ "}}" }}"
  }
  ANNOTATIONS {
    summary = "High invocation total on {{ "{{" }} $labels.instance {{ "}}" }}",
    description =  "High invocation total on {{ "{{" }} $labels.instance {{ "}}" }}"
  }
EOH

        destination   = "/etc/prometheus/alert.rules"
        change_mode   = "noop"
        change_signal = "SIGINT"
      }

      config {
        image = "prom/prometheus:v1.5.2"

        args = [
          "-config.file=/etc/prometheus/prometheus.yml",
          "-storage.local.path=/prometheus",
          "-storage.local.memory-chunks=10000",
          "--alertmanager.url=http://alertmanager.service.dc1.consul:9093",
        ]

        dns_servers = ["${NOMAD_IP_http}", "8.8.8.8", "8.8.8.4"]

        port_map {
          http = 9090
        }

        volumes = [
          "etc/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml",
          "etc/prometheus/alert.rules:/etc/prometheus/alert.rules",
        ]
      }

      resources {
        cpu    = 200 # 200 MHz
        memory = 256 # 256MB

        network {
          mbits = 10

          port "http" {
            static = 9090
          }
        }
      }

      service {
        port = "http"
        name = "prometheus"
        tags = ["faas"]

        check {
          type     = "http"
          port     = "http"
          interval = "10s"
          timeout  = "2s"
          path     = "/graph"
        }
      }
    }

    task "grafana" {
      driver = "docker"

      config {
        image = "grafana/grafana:4.5.2"

        port_map {
          http = 3000
        }
      }

      resources {
        cpu    = 200 # 500 MHz
        memory = 256 # 256MB

        network {
          mbits = 10

          port "http" {
            static = 3000
          }
        }
      }
    }
  }
}
