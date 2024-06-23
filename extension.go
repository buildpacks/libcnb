/*
 * Copyright 2023 the original author or authors.
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

package libcnb

// ExtensionInfo is information about the extension.
type ExtensionInfo struct {
	// ID is the ID of the extension.
	ID string `toml:"id"`

	// Name is the name of the extension.
	Name string `toml:"name"`

	// Version is the version of the extension.
	Version string `toml:"version"`

	// Homepage is the homepage of the extension.
	Homepage string `toml:"homepage"`

	// Description is a string describing the extension.
	Description string `toml:"description"`

	// Keywords is a list of words that are associated with the extension.
	Keywords []string `toml:"keywords"`

	// Licenses a list of extension licenses.
	Licenses []License `toml:"licenses"`
}

// Extension is the contents of the extension.toml file.
type Extension[EM any] struct {
	// API is the api version expected by the extension.
	API string `toml:"api"`

	// Info is information about the extension.
	Info ExtensionInfo `toml:"extension"`

	// Path is the path to the extension.
	Path string `toml:"-"`

	// Targets is the collection of targets supported by the buildpack.
	Targets []Target `toml:"targets"`

	// Metadata is arbitrary metadata attached to the extension.
	Metadata map[string]EM `toml:"metadata"`
}
