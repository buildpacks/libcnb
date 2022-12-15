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

	"github.com/BurntSushi/toml"
	"github.com/Masterminds/semver"

	"github.com/buildpacks/libcnb/internal"
)

// DetectContext contains the inputs to detection.
type DetectContext struct {

	// ApplicationPath is the location of the application source code as provided by
	// the lifecycle.
	ApplicationPath string

	// Buildpack is metadata about the buildpack, from buildpack.toml.
	Buildpack Buildpack

	// Logger is the way to write messages to the end user
	Logger Logger

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

// DetectFunc takes a context and returns a result, performing buildpack detect behaviors.
type DetectFunc func(context DetectContext) (DetectResult, error)

// Detect is called by the main function of a buildpack, for detection.
func Detect(detect DetectFunc, config Config) {
	var (
		err  error
		file string
		ok   bool
	)
	ctx := DetectContext{Logger: config.logger}

	ctx.ApplicationPath, err = os.Getwd()
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to get working directory\n%w", err))
		return
	}
	if config.logger.IsDebugEnabled() {
		config.logger.Debug(ApplicationPathFormatter(ctx.ApplicationPath))
	}

	if s, ok := os.LookupEnv(EnvBuildpackDirectory); ok {
		ctx.Buildpack.Path = filepath.Clean(s)
	} else {
		config.exitHandler.Error(fmt.Errorf("unable to get CNB_BUILDPACK_DIR, not found"))
		return
	}

	if config.logger.IsDebugEnabled() {
		config.logger.Debug(BuildpackPathFormatter(ctx.Buildpack.Path))
	}

	file = filepath.Join(ctx.Buildpack.Path, "buildpack.toml")
	if _, err = toml.DecodeFile(file, &ctx.Buildpack); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode buildpack %s\n%w", file, err))
		return
	}
	config.logger.Debugf("Buildpack: %+v", ctx.Buildpack)

	API, err := semver.NewVersion(ctx.Buildpack.API)
	if err != nil {
		config.exitHandler.Error(errors.New("version cannot be parsed"))
		return
	}

	compatVersionCheck, _ := semver.NewConstraint(fmt.Sprintf(">= %s, <= %s", MinSupportedBPVersion, MaxSupportedBPVersion))
	if !compatVersionCheck.Check(API) {
		if MinSupportedBPVersion == MaxSupportedBPVersion {
			config.exitHandler.Error(fmt.Errorf("this version of libcnb is only compatible with buildpack API == %s", MinSupportedBPVersion))
			return
		}

		config.exitHandler.Error(fmt.Errorf("this version of libcnb is only compatible with buildpack APIs >= %s, <= %s", MinSupportedBPVersion, MaxSupportedBPVersion))
		return
	}

	var buildPlanPath string

	ctx.Platform.Path, ok = os.LookupEnv(EnvPlatformDirectory)
	if !ok {
		config.exitHandler.Error(fmt.Errorf("expected CNB_PLATFORM_DIR to be set"))
		return
	}

	buildPlanPath, ok = os.LookupEnv(EnvDetectPlanPath)
	if !ok {
		config.exitHandler.Error(fmt.Errorf("expected CNB_BUILD_PLAN_PATH to be set"))
		return
	}

	if config.logger.IsDebugEnabled() {
		config.logger.Debug(PlatformFormatter(ctx.Platform))
	}

	file = filepath.Join(ctx.Platform.Path, "bindings")
	if ctx.Platform.Bindings, err = NewBindings(ctx.Platform.Path); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform bindings %s\n%w", file, err))
		return
	}
	config.logger.Debugf("Platform Bindings: %+v", ctx.Platform.Bindings)

	file = filepath.Join(ctx.Platform.Path, "env")
	if ctx.Platform.Environment, err = internal.NewConfigMapFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform environment %s\n%w", file, err))
		return
	}
	config.logger.Debugf("Platform Environment: %s", ctx.Platform.Environment)

	if ctx.StackID, ok = os.LookupEnv(EnvStackID); !ok {
		config.exitHandler.Error(fmt.Errorf("CNB_STACK_ID not set"))
		return
	}
	config.logger.Debugf("Stack: %s", ctx.StackID)

	result, err := detect(ctx)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
	config.logger.Debugf("Result: %+v", result)

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

		config.logger.Debugf("Writing build plans: %s <= %+v", buildPlanPath, plans)
		if err := config.tomlWriter.Write(buildPlanPath, plans); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write buildplan %s\n%w", buildPlanPath, err))
			return
		}
	}

	config.exitHandler.Pass()
}
