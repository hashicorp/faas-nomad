bind_addr = "0.0.0.0" # the default

data_dir  = "/tmp/nomad"

advertise {
  # Defaults to the node's hostname. If the hostname resolves to a loopback
  # address you must manually configure advertise addresses.
  http = "192.168.1.113"
  rpc  = "192.168.1.113"
  serf = "192.168.1.113:5648" # non-default ports may be specified
}

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled = true
  network_interface = "en0"
}

consul {
  address = "192.168.1.113:8500"
}
