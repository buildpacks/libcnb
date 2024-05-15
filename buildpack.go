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

// BuildpackInfo is information about the buildpack.
type BuildpackInfo struct {
	// ID is the ID of the buildpack.
	ID string `toml:"id"`

	// Name is the name of the buildpack.
	Name string `toml:"name"`

	// Version is the version of the buildpack.
	Version string `toml:"version"`

	// Homepage is the homepage of the buildpack.
	Homepage string `toml:"homepage"`

	// ClearEnvironment is whether the environment should be clear of user-configured environment variables.
	ClearEnvironment bool `toml:"clear-env"`

	// Description is a string describing the buildpack.
	Description string `toml:"description"`

	// Keywords is a list of words that are associated with the buildpack.
	Keywords []string `toml:"keywords"`

	// Licenses a list of buildpack licenses.
	Licenses []License `toml:"licenses"`

	// SBOM is the list of supported SBOM media types
	SBOMFormats []string `toml:"sbom-formats"`
}

// License contains information about a Software License
// governing the use or redistribution of a buildpack
type License struct {
	// Type is the identifier for the license.
	// It MAY use the SPDX 2.1 license expression, but is not limited to identifiers in the SPDX Licenses List.
	Type string `toml:"type"`

	// URI may be specified in lieu of or in addition to type to point to the license
	// if this buildpack is using a nonstandard license.
	URI string `toml:"uri"`
}

// BuildpackOrderBuildpack is a buildpack within in a buildpack order group.
type BuildpackOrderBuildpack struct {
	// ID is the id of the buildpack.
	ID string `toml:"id"`

	// Version is the version of the buildpack.
	Version string `toml:"version"`

	// Optional is whether the buildpack is optional within the buildpack.
	Optional bool `toml:"optional"`
}

// BuildpackOrder is an order definition in the buildpack.
type BuildpackOrder struct {
	// Groups is the collection of groups within the order.
	Groups []BuildpackOrderBuildpack `toml:"group"`
}

// Deprecated: BuildpackStack is a stack supported by the buildpack.
type BuildpackStack struct {
	// ID is the id of the stack.
	ID string `toml:"id"`

	// Mixins is the collection of mixins associated with the stack.
	Mixins []string `toml:"mixins"`
}

// BuildpackTargetDistro is the supported target distro
type BuildpackTargetDistro struct {
	// Name is the name of the supported distro.
	Name string `toml:"name"`

	// Version is the version of the supported distro.
	Version string `toml:"version"`
}

// BuildpackTarget is a target supported by the buildpack.
type BuildpackTarget struct {
	// OS is the supported os.
	OS string `toml:"os"`

	// Arch is the supported architecture.
	Arch string `toml:"arch"`

	// Variant is the supported variant of the architecture.
	Variant string `toml:"variant"`

	// Distros is the collection of distros associated with the target.
	Distros []BuildpackTargetDistro `toml:"distros"`
}

// Buildpack is the contents of the buildpack.toml file.
type Buildpack struct {
	// API is the api version expected by the buildpack.
	API string `toml:"api"`

	// Info is information about the buildpack.
	Info BuildpackInfo `toml:"buildpack"`

	// Path is the path to the buildpack.
	Path string `toml:"-"`

	// Deprecated: Stacks is the collection of stacks supported by the buildpack.
	Stacks []BuildpackStack `toml:"stacks"`

	// Targets is the collection of targets supported by the buildpack.
	Targets []BuildpackTarget `toml:"targets"`

	// Metadata is arbitrary metadata attached to the buildpack.
	Metadata map[string]interface{} `toml:"metadata"`
}
