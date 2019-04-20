{% set interface_address = '192.168.50.2' %}

include:
  - base
nomad:
  config:
    datacenter: dc1
    tls:
      http: True
      ca_file: /home/vagrant/placeholder-ca.crt
      cert_file: /home/vagrant/placeholder.crt
      key_file: /home/vagrant/placeholder.key
    advertise:
      http: {{ interface_address }}
    server:
      enabled: true
      bootstrap_expect: 1
      encrypt: "AaABbB+CcCdDdEeeFFfggG=="
    client:
      {% if grains['provider'] == 'virtualbox' %}
      network_interface: enp0s8
      {% elif grains['provider'] == 'vmware' %}
      network_interface: eth1
      {% elif grains['provider'] == 'libvirt' %}
      network_interface: eth0
      {% endif %}
      enabled: true
      meta:
        service_host: "true"
        faas_host: "true"
    consul:
      address: "127.0.0.1:8500"
      server_service_name: "nomad"
      client_service_name: "nomad-client"
      auto_advertise: true
      server_auto_join: true
      client_auto_join: true
    vault:
      enabled: true
      address: "http://127.0.0.1:8200"
      token: vagrant
  datacenters:
    - dc1
consul:
  config:
    connect:
      enabled: True
      ca_provider: vault
      ca_config:
        address: "http://127.0.0.1:8200"
        token: vagrant
        root_pki_path: pki
        intermediate_pki_path: pki_int
    server: True
    advertise_addr: {{ interface_address }}
    addresses:
      http: 0.0.0.0
      dns: 0.0.0.0
    ports:
      dns: 53
      http: 8500
    enable_debug: True
    datacenter: dc1
    encrypt: "RIxqpNlOXqtr/j4BgvIMEw=="
    bootstrap: true