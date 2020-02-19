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

package internal

import (
	"fmt"
	"io"
	"os"
)

const (
	// ErrorStatusCode is the status code returned for error.
	ErrorStatusCode = 1

	// FailStatusCode is the status code returned for fail.
	FailStatusCode = 100

	// PassStatusCode is the status code returned for pass.
	PassStatusCode = 0
)

// ExitHandler is the default implementation of the libcnb.ExitHandler interface.
type ExitHandler struct {
	exitFunc func(int)
	writer   io.Writer
}

// Option is a function for configuring an ExitHandler instance.
type Option func(handler ExitHandler) ExitHandler

// WithExitHandler creates an Option that configures the exit function.
func WithExitHandlerExitFunc(exitFunc func(int)) Option {
	return func(handler ExitHandler) ExitHandler {
		handler.exitFunc = exitFunc
		return handler
	}
}

// WithExitHandlerWriter creates an Option that configures the writer.
func WithExitHandlerWriter(writer io.Writer) Option {
	return func(handler ExitHandler) ExitHandler {
		handler.writer = writer
		return handler
	}
}

// NewExitHandler creates a new instance that calls os.Exit and writes to os.stderr.
func NewExitHandler(options ...Option) ExitHandler {
	h := ExitHandler{
		exitFunc: os.Exit,
		writer:   os.Stderr,
	}

	for _, option := range options {
		h = option(h)
	}

	return h
}

func (e ExitHandler) Error(err error) {
	_, _ = fmt.Fprintln(e.writer, err)
	e.exitFunc(ErrorStatusCode)
}

func (e ExitHandler) Fail() {
	e.exitFunc(FailStatusCode)
}

func (e ExitHandler) Pass() {
	e.exitFunc(PassStatusCode)
}
