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

// LaunchTOML represents the contents of launch.toml.
type LaunchTOML struct {
	// Labels is the collection of image labels contributed by the buildpack.
	Labels []Label `toml:"labels"`

	// Processes is the collection of process types contributed by the buildpack.
	Processes []Process `toml:"processes"`

	// Slices is the collection of slices contributed by the buildpack.
	Slices []Slice `toml:"slices"`
}

func (l LaunchTOML) isEmpty() bool {
	return len(l.Labels) == 0 && len(l.Processes) == 0 && len(l.Slices) == 0
}
