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

	"github.com/buildpacks/libcnb/internal"
)

// ApplicationPathFormatter is the formatter for an ApplicationPath.
type ApplicationPathFormatter string

func (a ApplicationPathFormatter) String() string {
	contents, err := internal.DirectoryContents{Path: string(a)}.Get()
	if err != nil {
		return fmt.Sprintf("Application contents: %s", err)
	}

	return fmt.Sprintf("Application contents: %s", contents)
}

// BuildpackPathFormatter is the formatter for a BuildpackPath.
type BuildpackPathFormatter string

func (b BuildpackPathFormatter) String() string {
	contents, err := internal.DirectoryContents{Path: string(b)}.Get()
	if err != nil {
		return fmt.Sprintf("Buildpack contents: %s", err)
	}

	return fmt.Sprintf("Buildpack contents: %s", contents)
}

// PlatformFormatter is the formatter for a Platform.
type PlatformFormatter Platform

func (p PlatformFormatter) String() string {
	contents, err := internal.DirectoryContents{Path: p.Path}.Get()
	if err != nil {
		return fmt.Sprintf("Platform contents: %s", err)
	}

	return fmt.Sprintf("Platform contents: %s", contents)
}
