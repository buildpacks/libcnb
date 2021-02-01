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

// BuildpackPlan represents a buildpack plan.
type BuildpackPlan struct {

	// Entries represents all the buildpack plan entries.
	Entries []BuildpackPlanEntry `toml:"entries,omitempty"`
}

// BuildpackPlanEntry represents an entry in the buildpack plan.
type BuildpackPlanEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`

	// Metadata is the metadata of the entry.  Optional.
	Metadata map[string]interface{} `toml:"metadata,omitempty"`
}

// UnmetPlanEntry denotes an unmet buildpack plan entry. When a buildpack returns an UnmetPlanEntry
// in the BuildResult, any BuildpackPlanEntry with a matching Name will be provided to subsequent
// providers.
type UnmetPlanEntry struct {
	// Name represents the name of the entry.
	Name string `toml:"name"`
}
