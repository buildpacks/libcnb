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
	"io/ioutil"
	"os"
	"path/filepath"
)

// EnvironmentWriter is a type used to write an environment to file filesystem.
type EnvironmentWriter struct{}

// Write creates the path directory, and creates a new file for each key with the value as the contents of each file.
func (w EnvironmentWriter) Write(path string, environment map[string]string) error {
	if len(environment) == 0 {
		return nil
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("unable to mkdir %s\n%w", path, err)
	}

	for key, value := range environment {
		f := filepath.Join(path, key)
		// #nosec
		if err := ioutil.WriteFile(f, []byte(value), 0644); err != nil {
			return fmt.Errorf("unable to write file %s\n%w", f, err)
		}
	}

	return nil
}
