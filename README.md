[![Build Status](https://travis-ci.org/hashicorp/faas-nomad.svg)](https://travis-ci.org/hashicorp/faas-nomad)  
[![Docker Repository on Quay](https://quay.io/repository/nicholasjackson/faas-nomad/status "Docker Repository on Quay")](https://quay.io/repository/nicholasjackson/faas-nomad)

# faas-nomad
Nomad plugin for [OpenFaas](https://github.com/openfaas/faas) 

# Running with Docker for Mac or Linux locally
1. Build the plugin `make build_docker`
1. Run nomad and consul in local mode ./startNomad.sh
1. Run OpenFaas `nomad run faas.hcl`
1. OpenFaaS Interface `open http://${HOST_IP}:8081`
1. Deploy function with values ...  
image: functions/nodeinfo:latest  
name: info  
handler: node main.js
