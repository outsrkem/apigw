package config

import (
	"fmt"
	"os"
)

var (
	Version   = "" // Application version information
	GoVersion = "" // Golang compiler version
	GitCommit = "" // Git commit hash identifier
)

// versions stores compilation and version metadata
type versions struct {
	AppVersion string
	GoVersion  string
	GitCommit  string
}

// newVersions constructs versions instance with given version parameters
func newVersions(appv, gov, commit string) (*versions, error) {
	v := &versions{
		AppVersion: appv,
		GoVersion:  gov,
		GitCommit:  commit,
	}
	return v, nil
}

// Print outputs version details to console and exits program normally
func (v *versions) Print(versions *versions) {
	fmt.Println("Version: ", versions.AppVersion)
	fmt.Println("Go Version: ", versions.GoVersion)
	fmt.Println("Git Commit: ", versions.GitCommit)
	os.Exit(0)
}
