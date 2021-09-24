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

package libcnb

// Application is the user contributed application to build.
type Application struct {

	// Path is the path to the application.
	Path string
}

// Label represents an image label.
type Label struct {

	// Key is the key of the label.
	Key string `toml:"key"`

	// Value is the value of the label.
	Value string `toml:"value"`
}

// Process represents metadata about a type of command that can be run.
type Process struct {

	// Type is the type of the process.
	Type string `toml:"type"`

	// Command is the command of the process.
	Command string `toml:"command"`

	// Arguments are arguments to the command.
	Arguments []string `toml:"args"`

	// Command is exec'd directly by the os (no profile.d scripts run)
	Direct bool `toml:"direct,omitempty"`

	// Default can be set to true to indicate that the process
	// type being defined should be the default process type for the app image.
	Default bool `toml:"default,omitempty"`
}

// Slice represents metadata about a slice.
type Slice struct {

	// Paths are the contents of the slice.
	Paths []string `toml:"paths"`
}

// LaunchTOML represents the contents of launch.toml.
type LaunchTOML struct {

	// Labels is the collection of image labels contributed by the buildpack.
	Labels []Label `toml:"labels"`

	// Processes is the collection of process types contributed by the buildpack.
	Processes []Process `toml:"processes"`

	// Slices is the collection of slices contributed by the buildpack.
	Slices []Slice `toml:"slices"`

	// BOM is a collection of entries for the bill of materials.
	BOM []BOMEntry `toml:"bom"`
}

func (l LaunchTOML) isEmpty() bool {
	return len(l.Labels) == 0 && len(l.Processes) == 0 && len(l.Slices) == 0 && len(l.BOM) == 0
}

// BuildTOML represents the contents of build.toml.
type BuildTOML struct {
	// BOM contains the build-time bill of materials.
	BOM []BOMEntry `toml:"bom"`

	// Unmet is a collection of buildpack plan entries that should be passed through to subsequent providers.
	Unmet []UnmetPlanEntry
}

func (b BuildTOML) isEmpty() bool {
	return len(b.BOM) == 0 && len(b.Unmet) == 0
}

// BOMEntry contains a bill of materials entry.
type BOMEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`

	// Metadata is the metadata of the entry.  Optional.
	Metadata map[string]interface{} `toml:"metadata,omitempty"`

	// Launch indicates whether the given entry is included in app image. If launch is true the entry
	// will be added to the app image Bill of Materials. Launch should be true if the entry describes
	// the contents of a launch layer or app layer.
	Launch bool `toml:"-"`

	// Build indicates whether the given entry is available at build time. If build is true the entry
	// will be added to the build Bill of Materials.
	Build bool `toml:"-"`
}

// Store represents the contents of store.toml
type Store struct {

	// Metadata represents the persistent metadata.
	Metadata map[string]interface{} `toml:"metadata"`
}
