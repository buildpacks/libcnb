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

//go:generate mockery --name EnvironmentWriter --case=underscore

// EnvironmentWriter is the interface implemented by a type that wants to serialize a map of environment variables to
// the file system.
type EnvironmentWriter interface {

	// Write is called with the path to a directory where the environment variables should be serialized to and the
	// environment variables to serialize to that directory.
	Write(dir string, environment map[string]string) error
}

//go:generate mockery --name ExitHandler --case=underscore

// ExitHandler is the interface implemented by a type that wants to handle exit behavior when a buildpack encounters an
// error.
type ExitHandler interface {

	// Error is called when an error is encountered.
	Error(error)

	// Fail is called when a buildpack fails.
	Fail()

	// Pass is called when a buildpack passes.
	Pass()
}

//go:generate mockery --name TOMLWriter --case=underscore

// TOMLWriter is the interface implemented by a type that wants to serialize an object to a TOML file.
type TOMLWriter interface {

	// Write is called with the path that a TOML file should be written to and the object to serialize to that file.
	Write(path string, value interface{}) error
}

//go:generate mockery --name ExecDWriter --case=underscore

// ExecDWriter is the interface implemented by a type that wants to write exec.d output to file descriptor 3.
type ExecDWriter interface {

	// Write is called with the map of environment value key value
	// pairs that will be written out
	Write(value map[string]string) error
}

// Config is an object that contains configurable properties for execution.
type Config struct {
	arguments         []string
	environmentWriter EnvironmentWriter
	exitHandler       ExitHandler
	tomlWriter        TOMLWriter
	execdWriter       ExecDWriter
}

// Option is a function for configuring a Config instance.
type Option func(config Config) Config

// WithArguments creates an Option that sets a collection of arguments.
func WithArguments(arguments []string) Option {
	return func(config Config) Config {
		config.arguments = arguments
		return config
	}
}

// WithEnvironmentWriter creates an Option that sets an EnvironmentWriter implementation.
func WithEnvironmentWriter(environmentWriter EnvironmentWriter) Option {
	return func(config Config) Config {
		config.environmentWriter = environmentWriter
		return config
	}
}

// WithExitHandler creates an Option that sets an ExitHandler implementation.
func WithExitHandler(exitHandler ExitHandler) Option {
	return func(config Config) Config {
		config.exitHandler = exitHandler
		return config
	}
}

// WithTOMLWriter creates an Option that sets a TOMLWriter implementation.
func WithTOMLWriter(tomlWriter TOMLWriter) Option {
	return func(config Config) Config {
		config.tomlWriter = tomlWriter
		return config
	}
}

// WithExecDWriter creates an Option that sets a ExecDWriter implementation.
func WithExecDWriter(execdWriter ExecDWriter) Option {
	return func(config Config) Config {
		config.execdWriter = execdWriter
		return config
	}
}
