package version

import (
	"fmt"
	"runtime"
)

var (
	gitMajor   string
	gitMinor   string
	gitURL     string
	gitVersion string
	gitCommit  string
	buildDate  string
)

type Info struct {
	Major      string `json:"major"`
	Minor      string `json:"minor"`
	GitURL     string `json:"gitURL"`
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	BuildDate  string `json:"buildDate"`
	GoVersion  string `json:"goVersion"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

func Get() Info {
	return Info{
		Major:      gitMajor,
		Minor:      gitMinor,
		GitURL:     gitURL,
		GitVersion: gitVersion,
		GitCommit:  gitCommit,
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
