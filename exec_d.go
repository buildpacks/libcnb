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

package libcnb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/v2/internal"
)

//go:generate mockery --name ExecD --case=underscore

// ExecD describes an interface for types that follow the Exec.d specification.
// It should return a map of environment variables and their values as output.
type ExecD interface {
	Execute() (map[string]string, error)
}

// RunExecD is called by the main function of a buildpack's execd binary, encompassing multiple execd
// executors in one binary.
func RunExecD(execDMap map[string]ExecD, options ...Option) {
	config := Config{
		arguments:   os.Args,
		execdWriter: internal.NewExecDWriter(),
		exitHandler: internal.NewExitHandler(),
	}

	for _, option := range options {
		config = option(config)
	}

	if len(config.arguments) == 0 {
		config.exitHandler.Error(fmt.Errorf("expected command name"))

		return
	}

	c := filepath.Base(config.arguments[0])
	e, ok := execDMap[c]
	if !ok {
		config.exitHandler.Error(fmt.Errorf("unsupported command %s", c))
		return
	}

	r, err := e.Execute()
	if err != nil {
		config.exitHandler.Error(err)
		return
	}

	if err := config.execdWriter.Write(r); err != nil {
		config.exitHandler.Error(err)
		return
	}
}
