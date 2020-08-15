package rundata

type ContainerEngine struct {
	Type   string
	Docker Docker
}

type Docker struct {
	Version        string
	CGroupDriver   string
	LogDriver      string
	LogOptsMaxSize string
	StorageDriver  string
}
