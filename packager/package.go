package packager

import (
	"fmt"
	"time"
)

type Package struct {
	Name          string    `json:"name"`
	Time          time.Time `json:"time"`
	Extension     string    `json:"extension"`
	ChunkSuffixes []string  `json:"chunkSuffixes"`
}

func NewPackage(name string, packageTime time.Time) *Package {
	return &Package{
		Name:      name,
		Time:      packageTime,
		Extension: "tar",
	}
}

func (p *Package) BaseName() string {
	return fmt.Sprintf("%s.%s", p.Name, p.Extension)
}

func (p *Package) FileNames() []string {
	if len(p.ChunkSuffixes) == 0 {
		return []string{p.BaseName()}
	}

	fileNames := []string{}
	baseName := p.BaseName()
	for i := range p.ChunkSuffixes {
		suffix := p.ChunkSuffixes[i]
		fileNames = append(fileNames, fmt.Sprintf("%s-%s", baseName, suffix))
	}
	return fileNames
}
