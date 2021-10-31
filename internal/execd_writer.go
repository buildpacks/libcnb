/*
 * Copyright 2018-2021 the original author or authors.
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
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

// ExecDWriter is a type used to write TOML files to fd3.
type ExecDWriter struct {
	outputWriter io.Writer
}

// Option is a function for configuring an ExitHandler instance.
type ExecDOption func(handler ExecDWriter) ExecDWriter

// WithExecDOutputWriter creates an Option that configures the writer.
func WithExecDOutputWriter(writer io.Writer) ExecDOption {
	return func(execdWriter ExecDWriter) ExecDWriter {
		execdWriter.outputWriter = writer
		return execdWriter
	}
}

// NewExitHandler creates a new instance that calls os.Exit and writes to os.stderr.
func NewExecDWriter(options ...ExecDOption) ExecDWriter {
	h := ExecDWriter{
		outputWriter: os.NewFile(3, "/dev/fd/3"),
	}

	for _, option := range options {
		h = option(h)
	}

	return h
}

// Write outputs the value serialized in TOML format to the appropriate writer.
func (e ExecDWriter) Write(value map[string]string) error {
	if value == nil {
		return nil
	}

	return toml.NewEncoder(e.outputWriter).Encode(value)
}
