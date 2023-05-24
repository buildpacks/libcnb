/*
 * Copyright 2018-2023 the original author or authors.
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
	"path/filepath"
)

func main(detect DetectFunc, build BuildFunc, generate GenerateFunc, options ...Option) {
	config := NewConfig(options...)

	if len(config.arguments) == 0 {
		config.exitHandler.Error(fmt.Errorf("expected command name"))
		return
	}

	switch c := filepath.Base(config.arguments[0]); c {
	case "build":
		Build(build, config)
	case "detect":
		Detect(detect, config)
	case "generate":
		Generate(generate, config)
	default:
		config.exitHandler.Error(fmt.Errorf("unsupported command %s", c))
		return
	}
}

// BuildpackMain is called by the main function of a buildpack, encapsulating both detection and build in the same binary.
func BuildpackMain(detect DetectFunc, build BuildFunc, options ...Option) {
	main(detect, build, nil, options...)
}

// ExtensionMain is called by the main function of a extension, encapsulating both detection and generation in the same binary.
func ExtensionMain(detect DetectFunc, generate GenerateFunc, options ...Option) {
	main(detect, nil, generate, options...)
}
