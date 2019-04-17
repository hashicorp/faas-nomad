package types

type NomadConfig struct {
	Address       string
	ACLToken      string
	TLSCA         string
	TLSCert       string
	TLSPrivateKey string
	TLSSkipVerify bool
}
