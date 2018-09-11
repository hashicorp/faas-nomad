{% set host_address = '192.168.50.1' %}

include:
  - base

faas_core_job_file:
  file.managed:
    - name: /tmp/faas.hcl
    - source: salt://nomad/files/faas.hcl
    - template: jinja
    - context:
      host_address: {{ host_address }}

faas_monitoring_job_file:
  file.managed:
    - name: /tmp/monitoring.hcl
    - source: salt://nomad/files/monitoring.hcl

faas_provider_file:
  file.managed:
    - name: /tmp/provider
    - content: {{ grains['provider'] }}