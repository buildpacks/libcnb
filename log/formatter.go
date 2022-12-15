/*
 * Copyright 2018-2022 the original author or authors.
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

package log

import (
	"os"
)

//go:generate mockery --name DirectoryContentFormatter --case=underscore

// DirectoryContentFormatter allows customization of logged directory output
//
// When libcnb logs the contents of a directory, each item in the directory
// is passed through a DirectoryContentFormatter.
//
// DirectoryContentsWriter implements this workflow:
//   - call RootPath(string) with the root path that's being walked
//   - call Title(string) with the given title, the output is logged
//   - for each file in the directory:
//   - call File(string, os.FileInfo), the output is logged
//
// # A default implementation is provided that returns a formatter applies no formatting
//
// The returned formatter operates as such:
//
//	Title -> returns string followed by `:\n`
//	File  -> returns file name relative to the root followed by `\n`
//
// A buildpack author could provide their own implementation through
// WithDirectoryContentFormatter when calling Detect or Build.
//
// A custom implementation might log in color or might log additional
// information about each file, like permissions. The implementation can
// also control line endings to force all of the files to be logged on a
// single line, or as multiple lines.
type DirectoryContentFormatter interface {
	// File takes the full path and os.FileInfo and returns a display string
	File(path string, info os.FileInfo) (string, error)

	// RootPath provides the root path being iterated
	RootPath(path string)

	// Title provides a plain string title which can be embellished
	Title(title string) string
}
