package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

type PlainDirectoryContentFormatter struct {
	rootPath string
}

// NewPlainDirectoryContentFormatter returns a formatter applies no formatting
//
// The returned formatter operates as such:
//
//	Title -> returns string followed by `:\n`
//	File  -> returns file name relative to the root followed by `\n`
func NewPlainDirectoryContentFormatter() *PlainDirectoryContentFormatter {
	return &PlainDirectoryContentFormatter{}
}

func (p *PlainDirectoryContentFormatter) File(path string, _ os.FileInfo) (string, error) {
	rel, err := filepath.Rel(p.rootPath, path)
	if err != nil {
		return "", fmt.Errorf("unable to calculate relative path %s -> %s\n%w", p.rootPath, path, err)
	}

	return fmt.Sprintf("%s\n", rel), nil
}

func (p *PlainDirectoryContentFormatter) RootPath(path string) {
	p.rootPath = path
}

func (p *PlainDirectoryContentFormatter) Title(title string) string {
	return fmt.Sprintf("%s:\n", title)
}
