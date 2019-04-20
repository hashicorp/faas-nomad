#!/bin/bash
echo 'Waiting for consul...'
while true
do
  START=`consul members | grep "alive"`
  if [ -n "$START" ]; then
    break
  else
    sleep 2
  fi
done

export NOMAD_ADDR=https://192.168.50.2:4646
export NOMAD_CACERT=/home/vagrant/placeholder-ca.crt
export NOMAD_CLIENT_CERT=/home/vagrant/placeholder.crt
export NOMAD_CLIENT_KEY=/home/vagrant/placeholder.key
echo 'Waiting for nomad...'
while true
do
  START=`nomad node-status | grep "ready"`
  if [ -n "$START" ]; then
    break
  else
    sleep 2
  fi
done
echo 'Deploying openfaas components...'
nomad run /tmp/faas.hcl
nomad run /tmp/monitoring.hcl