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
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/internal"
)

// Main is called by the main function of a buildpack, encapsulating both detection and build in the same binary.
func Main(detector Detector, builder Builder, options ...Option) {
	config := Config{
		arguments:         os.Args,
		environmentWriter: internal.EnvironmentWriter{},
		exitHandler:       internal.NewExitHandler(),
		tomlWriter:        internal.TOMLWriter{},
	}

	for _, option := range options {
		config = option(config)
	}

	if len(config.arguments) == 0 {
		config.exitHandler.Error(fmt.Errorf("expected command name"))
		return
	}

	switch c := filepath.Base(config.arguments[0]); c {
	case "build":
		Build(builder, options...)
	case "detect":
		Detect(detector, options...)
	default:
		config.exitHandler.Error(fmt.Errorf("unsupported command %s", c))
		return
	}
}
