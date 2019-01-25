package types

type ProviderConfig struct {
	VaultDefaultPolicy    string
	VaultSecretPathPrefix string
	VaultAppRoleID        string
	VaultAppSecretID      string
	Datacenter            string
	ConsulAddress         string
	ConsulDNSEnabled      bool
}
