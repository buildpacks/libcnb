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

// BuildPlanProvide represents a dependency provided by a buildpack.
type BuildPlanProvide struct {
	// Name is the name of the dependency.
	Name string `toml:"name"`
}

// BuildPlanRequire represents a dependency required by a buildpack.
type BuildPlanRequire[PL any] struct {
	// Name is the name of the dependency.
	Name string `toml:"name"`

	// Metadata is the metadata for the dependency. Optional.
	Metadata map[string]PL `toml:"metadata,omitempty"`
}

// BuildPlan represents the provisions and requirements of a buildpack during detection.
type BuildPlan[PL any] struct {
	// Provides is the dependencies provided by the buildpack.
	Provides []BuildPlanProvide `toml:"provides,omitempty"`

	// Requires is the dependencies required by the buildpack.
	Requires []BuildPlanRequire[PL] `toml:"requires,omitempty"`
}

// BuildPlans represents a collection of build plans produced by a buildpack during detection.
type BuildPlans[PL any] struct {
	// BuildPlan is the first build plan.
	BuildPlan[PL]

	// Or is the collection of other build plans.
	Or []BuildPlan[PL] `toml:"or,omitempty"`
}
