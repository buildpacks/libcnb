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

// Process represents metadata about a type of command that can be run.
type Process struct {
	// Type is the type of the process.
	Type string `toml:"type"`

	// Command is the command of the process.
	Command []string `toml:"command"`

	// Arguments are arguments to the command.
	Arguments []string `toml:"args"`

	// WorkingDirectory is a directory to execute the command in, removes the need to use a shell environment to CD into working directory
	WorkingDirectory string `toml:"working-dir,omitempty"`

	// Default can be set to true to indicate that the process
	// type being defined should be the default process type for the app image.
	Default bool `toml:"default,omitempty"`
}
