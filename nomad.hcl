bind_addr = "0.0.0.0" # the default

data_dir  = "/tmp/nomad"

advertise {
  # Defaults to the node's hostname. If the hostname resolves to a loopback
  # address you must manually configure advertise addresses.
  http = "192.168.0.101"
  rpc  = "192.168.0.101"
  serf = "192.168.0.101:5648" # non-default ports may be specified
}

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled = true
}

consul {
  address = "192.168.0.101:8500"
}
