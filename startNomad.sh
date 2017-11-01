#!/bin/bash

# Get the ip address
if [[ $OSTYPE == darwin* ]]; then
  IP_ADDRESS=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1')
else
  IP_ADDRESS=$(ip route get 1 | awk '{print $NF;exit}')
fi

# Set IP Address in config
sed "s/##HOST_IP##/${IP_ADDRESS}/g" < nomad_config.hcl.tmpl > nomad.hcl
echo "Discovered IP Address: ${IP_ADDRESS}"

# Create logs folder if needed
if [ ! -d "${HOME}/log" ]; then
  mkdir $HOME/log
fi

# Start Consul
echo "Starting Consul, redirecting logs to $HOME/log/consul.log"
sudo -b HOST_IP=${HOST_IP} nohup consul agent -dev -bind ${IP_ADDRESS} -dns-port 53 -client ${IP_ADDRESS} >~/log/consul.log 2>&1

# Start Nomad
echo "Starting Nomad, redirecting logs to $HOME/log/nomad.log"
nohup nomad agent --config=nomad.hcl >~/log/nomad.log 2>&1 &

# Set Nomad environment variable
echo ""
echo "You can set the following environment variables"
echo "export NOMAD_ADDR=http://${IP_ADDRESS}:4646"
echo "export CONSUL_HTTP_ADDR=http://${IP_ADDRESS}:8500"
echo "export FAAS_GATEWAY=http://${IP_ADDRESS}:8080"

