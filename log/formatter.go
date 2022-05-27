/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package log

import (
	"fmt"
	"os"
	"path/filepath"
)

//go:generate mockery --name DirectoryContentFormatter --case=underscore

// DirectoryContentFormatter allows customization of the directory output
type DirectoryContentFormatter interface {
	// File takes the full path and os.FileInfo and returns a display string
	File(path string, info os.FileInfo) (string, error)

	// RootPath provides the root path being iterated
	RootPath(path string)

	// Title provides a plain string title which can be embellished
	Title(title string) string
}

// PlainDirectoryContentFormatter applies no formatting
type PlainDirectoryContentFormatter struct {
	rootPath string
}

func NewPlainDirectoryContentFormatter() PlainDirectoryContentFormatter {
	return PlainDirectoryContentFormatter{}
}

func (p PlainDirectoryContentFormatter) File(path string, info os.FileInfo) (string, error) {
	rel, err := filepath.Rel(p.rootPath, path)
	if err != nil {
		return "", fmt.Errorf("unable to calculate relative path %s -> %s\n%w", p.rootPath, path, err)
	}

	return fmt.Sprintf("%s\n", rel), nil
}

func (p *PlainDirectoryContentFormatter) RootPath(path string) {
	p.rootPath = path
}

func (p PlainDirectoryContentFormatter) Title(title string) string {
	return fmt.Sprintf("%s:\n", title)
}
