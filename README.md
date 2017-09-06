# faas-nomad
Nomad plugin for OpenFaas

# Running with Docker for Mac
1. Build the plugin `make build_docker`
2. Start nomad `nomad agent -dev`
3. Run OpenFaas `nomad run faas.hcl`
4. Open FaaS Interface `open http://localhost:8081`
