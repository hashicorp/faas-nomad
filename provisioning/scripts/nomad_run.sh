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