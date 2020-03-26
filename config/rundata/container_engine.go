package rundata

type ContainerEngine struct {
	Type   int
	Docker Docker
}

type Docker struct {
	Version        string
	CGroupDriver   string
	LogDriver      string
	LogOptsMaxSize string
	StorageDriver  string
}
