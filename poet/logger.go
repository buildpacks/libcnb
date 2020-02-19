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

package poet

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Logger logs messages to a writer.
type Logger struct {
	debug io.Writer
	info  io.Writer
}

// Option is a function that configures a Logger.
type Option func(Logger) Logger

// WithDebug configures the debug Writer.
func WithDebug(writer io.Writer) Option {
	return func(logger Logger) Logger {
		logger.debug = writer
		return logger
	}
}

// NewLoggerWithOptions create a new instance of Logger.  It configures the Logger with options.
func NewLoggerWithOptions(writer io.Writer, options ...Option) Logger {
	l := Logger{
		info: writer,
	}

	for _, option := range options {
		l = option(l)
	}

	return l
}

// NewLogger creates a new instance of Logger.  It configures debug logging if $BP_DEBUG is set.
func NewLogger(writer io.Writer) Logger {
	var options []Option

	if _, ok := os.LookupEnv("BP_DEBUG"); ok {
		options = append(options, WithDebug(writer))
	}

	return NewLoggerWithOptions(writer, options...)
}

// Debug logs a message to the configured debug writer.
func (l Logger) Debug(format string, a ...interface{}) {
	if !l.IsDebugEnabled() {
		return
	}

	l.printf(l.debug, format, a...)
}

// DebugWriter returns the configured debug writer.
func (l Logger) DebugWriter() io.Writer {
	return l.debug
}

// IsDebugEnabled indicates whether debug logging is enabled.
func (l Logger) IsDebugEnabled() bool {
	return l.debug != nil
}

// Info logs a message to the configured info writer.
func (l Logger) Info(format string, a ...interface{}) {
	if !l.IsInfoEnabled() {
		return
	}

	l.printf(l.info, format, a...)
}

// InfoWriter returns the configured info writer.
func (l Logger) InfoWriter() io.Writer {
	return l.info
}

// IsInfoEnabled indicates whether info logging is enabled.
func (l Logger) IsInfoEnabled() bool {
	return l.info != nil
}

func (Logger) printf(writer io.Writer, format string, a ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}

	_, _ = fmt.Fprintf(writer, format, a...)
}
