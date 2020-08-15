package rundata

type Kubernetes struct {
	Version string
	Token   Token
}

type Token struct {
	Token          string
	CaCertHash     string
	CertificateKey string
}
