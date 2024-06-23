/*
 * Copyright 2018-2024 the original author or authors.
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

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	BOMFormatCycloneDXExtension = "cdx.json"
	BOMFormatSPDXExtension      = "spdx.json"
	BOMFormatSyftExtension      = "syft.json"
	BOMMediaTypeCycloneDX       = "application/vnd.cyclonedx+json"
	BOMMediaTypeSPDX            = "application/spdx+json"
	BOMMediaTypeSyft            = "application/vnd.syft+json"
	BOMUnknown                  = "unknown"
)

// Exec represents the exec.d layer location
type Exec struct {
	// Path is the path to the exec.d directory.
	Path string
}

// FilePath returns the fully qualified file path for a given name.
func (e Exec) FilePath(name string) string {
	return filepath.Join(e.Path, name)
}

// ProcessFilePath returns the fully qualified file path for a given name.
func (e Exec) ProcessFilePath(processType string, name string) string {
	return filepath.Join(e.Path, processType, name)
}

// BOMFormat indicates the format of the SBOM entry
type SBOMFormat int

const (
	CycloneDXJSON SBOMFormat = iota
	SPDXJSON
	SyftJSON
	UnknownFormat
)

func (b SBOMFormat) String() string {
	return []string{
		BOMFormatCycloneDXExtension,
		BOMFormatSPDXExtension,
		BOMFormatSyftExtension,
		BOMUnknown}[b]
}

func (b SBOMFormat) MediaType() string {
	return []string{
		BOMMediaTypeCycloneDX,
		BOMMediaTypeSPDX,
		BOMMediaTypeSyft,
		BOMUnknown}[b]
}

func SBOMFormatFromString(from string) (SBOMFormat, error) {
	switch from {
	case CycloneDXJSON.String():
		return CycloneDXJSON, nil
	case SPDXJSON.String():
		return SPDXJSON, nil
	case SyftJSON.String():
		return SyftJSON, nil
	}

	return UnknownFormat, fmt.Errorf("unable to translate from %s to SBOMFormat", from)
}

// Contribute represents a layer managed by the buildpack.
type Layer[LM any] struct {
	// LayerTypes indicates the type of layer
	LayerTypes `toml:"types"`

	// Metadata is the metadata associated with the layer.
	Metadata map[string]LM `toml:"metadata"`

	// Name is the name of the layer.
	Name string `toml:"-"`

	// Path is the filesystem location of the layer.
	Path string `toml:"-"`

	// BuildEnvironment are the environment variables set at build time.
	BuildEnvironment Environment `toml:"-"`

	// LaunchEnvironment are the environment variables set at launch time.
	LaunchEnvironment Environment `toml:"-"`

	// SharedEnvironment are the environment variables set at both build and launch times.
	SharedEnvironment Environment `toml:"-"`

	// Exec is the exec.d executables set in the layer.
	Exec Exec `toml:"-"`
}

func (l Layer[LM]) Reset() (Layer[LM], error) {
	l.LayerTypes = LayerTypes{
		Build:  false,
		Launch: false,
		Cache:  false,
	}

	l.SharedEnvironment = Environment{}
	l.BuildEnvironment = Environment{}
	l.LaunchEnvironment = Environment{}
	l.Metadata = nil

	err := os.RemoveAll(l.Path)
	if err != nil {
		return Layer[LM]{}, fmt.Errorf("error could not remove file: %s", err)
	}

	err = os.MkdirAll(l.Path, os.ModePerm)
	if err != nil {
		return Layer[LM]{}, fmt.Errorf("error could not create directory: %s", err)
	}

	return l, nil
}

// SBOMPath returns the path to the layer specific SBOM File
func (l Layer[LM]) SBOMPath(bt SBOMFormat) string {
	return filepath.Join(filepath.Dir(l.Path), fmt.Sprintf("%s.sbom.%s", l.Name, bt))
}

// LayerTypes describes which types apply to a given layer. A layer may have any combination of Launch, Build, and
// Cache types.
type LayerTypes struct {
	// Build indicates that a layer should be used for builds.
	Build bool `toml:"build"`

	// Cache indicates that a layer should be cached.
	Cache bool `toml:"cache"`

	// Launch indicates that a layer should be used for launch.
	Launch bool `toml:"launch"`
}

// Layers represents the layers part of the specification.
type Layers[LM any] struct {
	// Path is the layers filesystem location.
	Path string
}

// Layer creates a new layer, loading metadata if it exists.
func (l *Layers[LM]) Layer(name string) (Layer[LM], error) {
	layer := Layer[LM]{
		Name:              name,
		Path:              filepath.Join(l.Path, name),
		BuildEnvironment:  Environment{},
		LaunchEnvironment: Environment{},
		SharedEnvironment: Environment{},
		Exec:              Exec{Path: filepath.Join(l.Path, name, "exec.d")},
	}

	f := filepath.Join(l.Path, fmt.Sprintf("%s.toml", name))
	if _, err := toml.DecodeFile(f, &layer); err != nil && !os.IsNotExist(err) {
		return Layer[LM]{}, fmt.Errorf("unable to decode layer metadata %s\n%w", f, err)
	}

	return layer, nil
}

// BOMBuildPath returns the full path to the build SBoM file for the buildpack
func (l Layers[LM]) BuildSBOMPath(bt SBOMFormat) string {
	return filepath.Join(l.Path, fmt.Sprintf("build.sbom.%s", bt))
}

// BOMLaunchPath returns the full path to the launch SBoM file for the buildpack
func (l Layers[LM]) LaunchSBOMPath(bt SBOMFormat) string {
	return filepath.Join(l.Path, fmt.Sprintf("launch.sbom.%s", bt))
}
