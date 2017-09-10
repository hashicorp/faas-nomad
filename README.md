[![Build Status](https://travis-ci.org/hashicorp/faas-nomad.svg)](https://travis-ci.org/hashicorp/faas-nomad)

# faas-nomad
Nomad plugin for [OpenFaas](https://github.com/alexellis/faas) 

# Running with Docker for Mac
1. Build the plugin `make build_docker`
1. Start nomad `nohup nomad agent -dev >./nomad.log 2>&1 &`
1. Start consul `nohup consul agent -dev >./consul.log 2>&1 &`
1. Run OpenFaas `nomad run faas.hcl`
1. OpenFaaS Interface `open http://localhost:8081`
