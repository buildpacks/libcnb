/*
 * Copyright 2018-2020 the original author or authors.
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

package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// DirectoryContents is used to generate a collection of the names of all files within a directory.
type DirectoryContents struct {
	Path string
}

// Get returns the names of all files within a directory
func (d DirectoryContents) Get() ([]string, error) {
	var contents []string

	if err := filepath.Walk(d.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(d.Path, path)
		if err != nil {
			return fmt.Errorf("unable to calculate relative path %s -> %s: %w", d.Path, path, err)
		}

		contents = append(contents, rel)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("error walking path %s: %w", d.Path, err)
	}

	sort.Strings(contents)
	return contents, nil
}
