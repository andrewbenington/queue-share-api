package version

import (
	"log"
	"runtime"
)

type GitVersion struct {
	Tag    string           `yaml:"tag" json:"tag"`
	Commit string           `yaml:"commit" json:"commit"`
	Tree   WorkingTreeState `yaml:"working_tree" json:"working_tree"`
}

type GoMetadata struct {
	Version string `yaml:"version" json:"version"`
	Arch    string `yaml:"arch" json:"arch"`
	OS      string `yaml:"os" json:"os"`
}

type Version struct {
	Git      GitVersion `yaml:"git" json:"git"`
	Database string     `yaml:"database" json:"database"`
	Go       GoMetadata `yaml:"go" json:"go"`
	Date     string     `yaml:"build_date" json:"build_date"`
}

type WorkingTreeState string

const (
	TREE_CLEAN WorkingTreeState = "clean"
	TREE_DIRTY WorkingTreeState = "dirty"
)

var (
	// Git tag
	Tag string
	// Git commit
	Commit string
	// Working tree state
	Tree string
	// Database version
	Database string
	// Go architecture
	Arch = runtime.GOARCH
	// Go version
	Go = runtime.Version()
	// Build OS
	OS = runtime.GOOS
	// Build date
	Date string
)

func Get() *Version {
	// check if a semantic tag was provided
	if len(Tag) == 0 {
		log.Print("no semantic tag provided - defaulting to v0.0.0")

		// set a fallback default for the tag
		Tag = "v0.0.0"
	}

	return &Version{
		Git: GitVersion{
			Tag:    Tag,
			Commit: Commit,
			Tree:   WorkingTreeState(Tree),
		},
		Database: Database,
		Go: GoMetadata{
			Version: Go,
			Arch:    Arch,
			OS:      OS,
		},
		Date: Date,
	}
}
