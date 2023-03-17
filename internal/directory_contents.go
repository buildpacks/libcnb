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

package internal

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/log"
)

// DirectoryContentsWriter is used write the contents of a directory to the given io.Writer
type DirectoryContentsWriter struct {
	format log.DirectoryContentFormatter
	writer io.Writer
}

// NewDirectoryContentsWriter returns a new DirectoryContentsWriter initialized and ready to be used
func NewDirectoryContentsWriter(format log.DirectoryContentFormatter, writer io.Writer) DirectoryContentsWriter {
	return DirectoryContentsWriter{
		format: format,
		writer: writer,
	}
}

// Write all the file contents to the writer
func (d DirectoryContentsWriter) Write(title, path string) error {
	d.format.RootPath(path)
	d.writer.Write([]byte(d.format.Title(title)))

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		msg, err := d.format.File(path, info)
		if err != nil {
			return fmt.Errorf("unable to format\n%w", err)
		}

		d.writer.Write([]byte(msg))

		return nil
	}); err != nil {
		return fmt.Errorf("error walking path %s\n%w", path, err)
	}

	return nil
}
