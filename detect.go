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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/buildpacks/libcnb/internal"
	"github.com/buildpacks/libcnb/poet"
)

// DetectContext contains the inputs to detection.
type DetectContext struct {

	// Application is the application to build.
	Application Application

	// Buildpack is metadata about the buildpack, from buildpack.toml.
	Buildpack Buildpack

	// Platform is the contents of the platform.
	Platform Platform

	// StackID is the ID of the stack.
	StackID string
}

// DetectResult contains the results of detection.
type DetectResult struct {

	// Pass indicates whether detection has passed.
	Pass bool

	// Plans are the build plans contributed by the buildpack.
	Plans []BuildPlan
}

//go:generate mockery --name Detector --case=underscore

// DetectFunc takes a context and returns a result, performing buildpack detect behaviors.
type DetectFunc func(context DetectContext) (DetectResult, error)

// Detect is called by the main function of a buildpack, for detection.
func Detect(detect DetectFunc, options ...Option) {
	config := Config{
		arguments:         os.Args,
		environmentWriter: internal.EnvironmentWriter{},
		exitHandler:       internal.NewExitHandler(),
		tomlWriter:        internal.TOMLWriter{},
	}

	for _, option := range options {
		config = option(config)
	}

	if len(config.arguments) != 3 {
		config.exitHandler.Error(fmt.Errorf("expected 2 arguments and received %d", len(config.arguments)-1))
		return
	}

	var (
		err  error
		file string
		ok   bool
	)
	ctx := DetectContext{}
	logger := poet.NewLogger(os.Stdout)

	ctx.Application.Path, err = os.Getwd()
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to get working directory\n%w", err))
		return
	}
	if logger.IsDebugEnabled() {
		logger.Debug(ApplicationPathFormatter(ctx.Application.Path))
	}

	if s, ok := os.LookupEnv("CNB_BUILDPACK_DIR"); ok {
		ctx.Buildpack.Path = filepath.Clean(s)
	} else { // TODO: Remove branch once lifecycle has been updated to support this
		ctx.Buildpack.Path = filepath.Clean(strings.TrimSuffix(config.arguments[0], filepath.Join("bin", "detect")))
	}
	if logger.IsDebugEnabled() {
		logger.Debug(BuildpackPathFormatter(ctx.Buildpack.Path))
	}

	file = filepath.Join(ctx.Buildpack.Path, "buildpack.toml")
	if _, err = toml.DecodeFile(file, &ctx.Buildpack); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode buildpack %s\n%w", file, err))
		return
	}
	logger.Debugf("Buildpack: %+v", ctx.Buildpack)

	API := strings.TrimSpace(ctx.Buildpack.API)
	if API != "0.5" && API != "0.6" {
		config.exitHandler.Error(errors.New("this version of libcnb is only compatible with buildpack API 0.5 and 0.6"))
		return
	}

	ctx.Platform.Path = config.arguments[1]
	if logger.IsDebugEnabled() {
		logger.Debug(PlatformFormatter(ctx.Platform))
	}

	file = filepath.Join(ctx.Platform.Path, "bindings")
	if ctx.Platform.Bindings, err = NewBindingsFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform bindings %s\n%w", file, err))
		return
	}
	logger.Debugf("Platform Bindings: %+v", ctx.Platform.Bindings)

	file = filepath.Join(ctx.Platform.Path, "env")
	if ctx.Platform.Environment, err = internal.NewConfigMapFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform environment %s\n%w", file, err))
		return
	}
	logger.Debugf("Platform Environment: %s", ctx.Platform.Environment)

	if ctx.StackID, ok = os.LookupEnv("CNB_STACK_ID"); !ok {
		config.exitHandler.Error(fmt.Errorf("CNB_STACK_ID not set"))
		return
	}
	logger.Debugf("Stack: %s", ctx.StackID)

	result, err := detect(ctx)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
	logger.Debugf("Result: %+v", result)

	if !result.Pass {
		config.exitHandler.Fail()
		return
	}

	if len(result.Plans) > 0 {
		var plans BuildPlans
		if len(result.Plans) > 0 {
			plans.BuildPlan = result.Plans[0]
		}
		if len(result.Plans) > 1 {
			plans.Or = result.Plans[1:]
		}

		file = config.arguments[2]
		logger.Debugf("Writing build plans: %s <= %+v", file, plans)
		if err := config.tomlWriter.Write(file, plans); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write buildplan %s\n%w", file, err))
			return
		}
	}

	config.exitHandler.Pass()
}
