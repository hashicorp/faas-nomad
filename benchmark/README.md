# Benchmarking faas-nomadd
To bench mark a function for testing things like scaling you can use the bench tool.

1. Install bench `go get github.com/nicholasjackson/bench/bench`
1. Change the address of the function to test in faas.go
1. Build the plugin `go build faas.go`
1. Run bench `$bench run -duration 100s -plugin ./faas -thread 10`
