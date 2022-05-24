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

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/buildpacks/libcnb/internal"
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

// Profile is the collection of values to be written into profile.d
type Profile map[string]string

// Add formats using the default formats for its operands and adds an entry for a .profile.d file. Spaces are added
// between operands when neither is a string.
func (p Profile) Add(name string, a ...interface{}) {
	p[name] = fmt.Sprint(a...)
}

// Addf formats according to a format specifier and adds an entry for a .profile.d file.
func (p Profile) Addf(name string, format string, a ...interface{}) {
	p[name] = fmt.Sprintf(format, a...)
}

// ProcessAdd formats using the default formats for its operands and adds an entry for a .profile.d file. Spaces are
// added between operands when neither is a string.
func (p Profile) ProcessAdd(processType string, name string, a ...interface{}) {
	p.Add(filepath.Join(processType, name), a...)
}

// ProcessAddf formats according to a format specifier and adds an entry for a .profile.d file.
func (p Profile) ProcessAddf(processType string, name string, format string, a ...interface{}) {
	p.Addf(filepath.Join(processType, name), format, a...)
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
type Layer struct {
	// LayerTypes indicates the type of layer
	LayerTypes `toml:"types"`

	// Metadata is the metadata associated with the layer.
	Metadata map[string]interface{} `toml:"metadata"`

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

	// Profile is the profile.d scripts set in the layer.
	Profile Profile `toml:"-"`

	// Exec is the exec.d executables set in the layer.
	Exec Exec `toml:"-"`
}

func (l Layer) Reset() (Layer, error) {
	l.LayerTypes = LayerTypes{
		Build:  false,
		Launch: false,
		Cache:  false,
	}

	l.SharedEnvironment = Environment{}
	l.BuildEnvironment = Environment{}
	l.LaunchEnvironment = Environment{}
	l.Profile = Profile{}
	l.Metadata = nil

	err := os.RemoveAll(l.Path)
	if err != nil {
		return Layer{}, fmt.Errorf("error could not remove file: %s", err)
	}

	err = os.MkdirAll(l.Path, os.ModePerm)
	if err != nil {
		return Layer{}, fmt.Errorf("error could not create directory: %s", err)
	}

	return l, nil
}

// SBOMPath returns the path to the layer specific SBOM File
func (l Layer) SBOMPath(bt SBOMFormat) string {
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
type Layers struct {
	// Path is the layers filesystem location.
	Path string
}

// Layer creates a new layer, loading metadata if it exists.
func (l *Layers) Layer(name string) (Layer, error) {
	layer := Layer{
		Name:              name,
		Path:              filepath.Join(l.Path, name),
		BuildEnvironment:  Environment{},
		LaunchEnvironment: Environment{},
		SharedEnvironment: Environment{},
		Profile:           Profile{},
		Exec:              Exec{Path: filepath.Join(l.Path, name, "exec.d")},
	}

	f := filepath.Join(l.Path, fmt.Sprintf("%s.toml", name))
	if _, err := toml.DecodeFile(f, &layer); err != nil && !os.IsNotExist(err) {
		return Layer{}, fmt.Errorf("unable to decode layer metadata %s\n%w", f, err)
	}

	if !layer.Build && !layer.Cache && !layer.Launch {
		// if all three are false, that could mean we have a API <= 0.5 TOML file
		// try parsing the <= 0.5 API format where these were top level attributes
		buf := internal.LayerAPI5{}
		if _, err := toml.DecodeFile(f, &buf); err != nil && !os.IsNotExist(err) {
			return Layer{}, fmt.Errorf("unable to decode layer metadata %s\n%w", f, err)
		}
		layer.Build = buf.Build
		layer.Cache = buf.Cache
		layer.Launch = buf.Launch
	}

	return layer, nil
}

// BOMBuildPath returns the full path to the build SBoM file for the buildpack
func (l Layers) BuildSBOMPath(bt SBOMFormat) string {
	return filepath.Join(l.Path, fmt.Sprintf("build.sbom.%s", bt))
}

// BOMLaunchPath returns the full path to the launch SBoM file for the buildpack
func (l Layers) LaunchSBOMPath(bt SBOMFormat) string {
	return filepath.Join(l.Path, fmt.Sprintf("launch.sbom.%s", bt))
}
