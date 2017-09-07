[![CircleCI](https://circleci.com/gh/nicholasjackson/faas-nomad.svg?style=svg)](https://circleci.com/gh/nicholasjackson/faas-nomad)

# faas-nomad
Nomad plugin for [OpenFaas](https://github.com/alexellis/faas) 

# Running with Docker for Mac
1. Build the plugin `make build_docker`
2. Start nomad `nomad agent -dev`
3. Run OpenFaas `nomad run faas.hcl`
4. OpenFaaS Interface `open http://localhost:8081`
