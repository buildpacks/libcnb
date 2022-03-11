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

package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// PlainLogger implements Logger and logs messages to a writer.
type PlainLogger struct {
	debug io.Writer
}

// New creates a new instance of PlainLogger.  It configures debug logging if $BP_DEBUG or $BP_LOG_LEVEL are set.
func New(debug io.Writer) PlainLogger {
	if strings.ToLower(os.Getenv("BP_LOG_LEVEL")) == "debug" || os.Getenv("BP_DEBUG") != "" {
		return PlainLogger{debug: debug}
	}

	return PlainLogger{}
}

// NewDiscard creates a new instance of PlainLogger that discards all log messages. Useful in testing.
func NewDiscard() PlainLogger {
	return PlainLogger{debug: io.Discard}
}

// Debug formats using the default formats for its operands and writes to the configured debug writer. Spaces are added
// between operands when neither is a string.
func (l PlainLogger) Debug(a ...interface{}) {
	if !l.IsDebugEnabled() {
		return
	}

	s := fmt.Sprint(a...)

	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}

	_, _ = fmt.Fprint(l.debug, s)
}

// Debugf formats according to a format specifier and writes to the configured debug writer.
func (l PlainLogger) Debugf(format string, a ...interface{}) {
	if !l.IsDebugEnabled() {
		return
	}

	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}

	_, _ = fmt.Fprintf(l.debug, format, a...)
}

// DebugWriter returns the configured debug writer.
func (l PlainLogger) DebugWriter() io.Writer {
	return l.debug
}

// IsDebugEnabled indicates whether debug logging is enabled.
func (l PlainLogger) IsDebugEnabled() bool {
	return l.debug != nil
}
