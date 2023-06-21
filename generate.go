/*
 * Copyright 2023 the original author or authors.
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
	"github.com/buildpacks/libcnb/log"
)

// GenerateContext contains the inputs to generate.
type GenerateContext struct {
	// ApplicationPath is the location of the application source code as provided by
	// the lifecycle.
	ApplicationPath string

	// Extension is metadata about the extension, from extension.toml.
	Extension Extension

	// OutputDirectory is the location Dockerfiles should be written to.
	OutputDirectory string

	// Logger is the way to write messages to the end user
	Logger log.Logger

	// Plan is the buildpack plan provided to the buildpack.
	Plan BuildpackPlan

	// Platform is the contents of the platform.
	Platform Platform

	// StackID is the ID of the stack.
	StackID string
}

// GenerateResult contains the results of detection.
type GenerateResult struct {
	// Unmet contains buildpack plan entries that were not satisfied by the buildpack and therefore should be
	// passed to subsequent providers.
	Unmet []UnmetPlanEntry
}

// NewBuildResult creates a new BuildResult instance, initializing empty fields.
func NewGenerateResult() GenerateResult {
	return GenerateResult{}
}

func (b GenerateResult) String() string {
	return fmt.Sprintf(
		"{Unmet:%+v}",
		b.Unmet,
	)
}

// BuildFunc takes a context and returns a result, performing extension generate behaviors.
type GenerateFunc func(context GenerateContext) (GenerateResult, error)

// Generate is called by the main function of a extension, for generate phase
func Generate(generate GenerateFunc, config Config) {
	var (
		err  error
		file string
		ok   bool
	)
	ctx := GenerateContext{Logger: config.logger}

	ctx.ApplicationPath, err = os.Getwd()
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to get working directory\n%w", err))
		return
	}

	if config.logger.IsDebugEnabled() {
		if err := config.contentWriter.Write("Application contents", ctx.ApplicationPath); err != nil {
			config.logger.Debugf("unable to write application contents\n%w", err)
		}
	}

	if s, ok := os.LookupEnv(EnvExtensionDirectory); ok {
		ctx.Extension.Path = filepath.Clean(s)
	} else {
		config.exitHandler.Error(fmt.Errorf("unable to get CNB_EXTENSION_DIR, not found"))
		return
	}

	if config.logger.IsDebugEnabled() {
		if err := config.contentWriter.Write("Extension contents", ctx.Extension.Path); err != nil {
			config.logger.Debugf("unable to write extension contents\n%w", err)
		}
	}

	file = filepath.Join(ctx.Extension.Path, "extension.toml")
	if _, err = toml.DecodeFile(file, &ctx.Extension); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode extension %s\n%w", file, err))
		return
	}
	config.logger.Debugf("Extension: %+v", ctx.Extension)

	API, err := semver.NewVersion(ctx.Extension.API)
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

	outputDir, ok := os.LookupEnv(EnvOutputDirectory)
	if !ok {
		config.exitHandler.Error(fmt.Errorf("expected CNB_OUTPUT_DIR to be set"))
		return
	}
	ctx.OutputDirectory = outputDir

	ctx.Platform.Path, ok = os.LookupEnv(EnvPlatformDirectory)
	if !ok {
		config.exitHandler.Error(fmt.Errorf("expected CNB_PLATFORM_DIR to be set"))
		return
	}

	buildpackPlanPath, ok := os.LookupEnv(EnvBuildPlanPath)
	if !ok {
		config.exitHandler.Error(fmt.Errorf("expected CNB_BP_PLAN_PATH to be set"))
		return
	}

	if config.logger.IsDebugEnabled() {
		if err := config.contentWriter.Write("Platform contents", ctx.Platform.Path); err != nil {
			config.logger.Debugf("unable to write platform contents\n%w", err)
		}
	}

	if ctx.Platform.Bindings, err = NewBindings(ctx.Platform.Path); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform bindings %s\n%w", ctx.Platform.Path, err))
		return
	}
	config.logger.Debugf("Platform Bindings: %+v", ctx.Platform.Bindings)

	file = filepath.Join(ctx.Platform.Path, "env")
	if ctx.Platform.Environment, err = internal.NewConfigMapFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform environment %s\n%w", file, err))
		return
	}
	config.logger.Debugf("Platform Environment: %s", ctx.Platform.Environment)

	if _, err = toml.DecodeFile(buildpackPlanPath, &ctx.Plan); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode buildpack plan %s\n%w", buildpackPlanPath, err))
		return
	}
	config.logger.Debugf("Buildpack Plan: %+v", ctx.Plan)

	if ctx.StackID, ok = os.LookupEnv(EnvStackID); !ok {
		config.exitHandler.Error(fmt.Errorf("CNB_STACK_ID not set"))
		return
	}
	config.logger.Debugf("Stack: %s", ctx.StackID)

	result, err := generate(ctx)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
	config.logger.Debugf("Result: %+v", result)
}
