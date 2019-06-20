package types

type ProviderConfig struct {
	Vault             VaultConfig
	Datacenter        string
	ConsulAddress     string
	ConsulDNSEnabled  bool
	CPUArchConstraint string
}
