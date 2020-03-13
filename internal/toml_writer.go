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

	"github.com/BurntSushi/toml"
)

// TOMLWriter is a type used to write TOML files to the filesystem.
type TOMLWriter struct{}

// Write creates the path's parent directories, and creates a new file or truncates an existing file and then marshals
// the value to the file.
func (TOMLWriter) Write(path string, value interface{}) error {
	if value == nil {
		return nil
	}

	d := filepath.Dir(path)
	if err := os.MkdirAll(d, 0755); err != nil {
		return fmt.Errorf("unable to mkdir %s\n%w", d, err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file %s\n%w", path, err)
	}
	defer file.Close()

	return toml.NewEncoder(file).Encode(value)
}
