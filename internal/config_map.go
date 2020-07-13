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
	"strings"
)

// ConfigMap represents a file-based projection of a collection of key-value pairs.
type ConfigMap map[string]string

// NewConfigMapFromPath creates a new ConfigMap from the files located within a given path.
func NewConfigMapFromPath(path string) (ConfigMap, error) {
	files, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, fmt.Errorf("unable to glob %s\n%w", path, err)
	}

	configMap := ConfigMap{}
	for _, file := range files {
		if strings.HasPrefix(filepath.Base(file), ".") {
			// ignore hidden files
			continue
		}
		if stat, err := os.Stat(file); err != nil {
			return nil, fmt.Errorf("failed to stat file %s\n%w", file, err)
		} else if stat.IsDir() {
			continue
		}
		contents, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("unable to read file %s\n%w", file, err)
		}

		configMap[filepath.Base(file)] = strings.TrimSpace(string(contents))
	}

	return configMap, nil
}
