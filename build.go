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
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb/internal"
	"github.com/buildpacks/libcnb/poet"
)

// BuildContext contains the inputs to build.
type BuildContext struct {

	// ApplicationPath is the path to the application.
	ApplicationPath string

	// Buildpack is metadata about the buildpack, from buildpack.toml.
	Buildpack Buildpack

	// BuildpackPath is the path to the buildpack.
	BuildpackPath string

	// Layers is the layers available to the buildpack.
	Layers Layers

	// PersistentMetadata is metadata that is persisted even across cache cleaning.
	PersistentMetadata map[string]interface{}

	// Plan is the buildpack plan provided to the buildpack.
	Plan BuildpackPlan

	// Platform is the contents of the platform.
	Platform Platform

	// StackID is the ID of the stack.
	StackID string
}

// BuildResult contains the results of detection.
type BuildResult struct {

	// Layers is the collection of LayerCreators contributed by the buildpack.
	Layers []LayerContributor

	// PersistentMetadata is metadata that is persisted even across cache cleaning.
	PersistentMetadata map[string]interface{}

	// Plan is the buildpack plan contributed by the buildpack.
	Plan BuildpackPlan

	// Processes are the process types contributed by the buildpack.
	Processes []Process

	// Slices are the application slices contributed by the buildpack.
	Slices []Slice
}

// BuildFunc is the callback function for buildpack build implementations.
type BuildFunc func(BuildContext) (BuildResult, error)

// Build is called by the main function of a buildpack, for build.
func Build(f BuildFunc, options ...Option) {
	config := Config{
		arguments:         os.Args,
		environmentWriter: internal.EnvironmentWriter{},
		exitHandler:       internal.NewExitHandler(),
		tomlWriter:        internal.TOMLWriter{},
	}

	for _, option := range options {
		config = option(config)
	}

	if len(config.arguments) != 4 {
		config.exitHandler.Error(fmt.Errorf("expected 3 arguments and received %d", len(config.arguments)-1))
		return
	}

	var (
		err  error
		file string
		ok   bool
	)
	ctx := BuildContext{}
	logger := poet.NewLogger(os.Stdout)

	ctx.ApplicationPath, err = os.Getwd()
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to get working directory: %w", err))
		return
	}
	if logger.IsDebugEnabled() {
		logger.Debug("%s", ApplicationPathFormatter(ctx.ApplicationPath))
	}

	ctx.BuildpackPath = filepath.Clean(strings.TrimSuffix(config.arguments[0], filepath.Join("bin", "build")))
	if logger.IsDebugEnabled() {
		logger.Debug("%s", BuildpackPathFormatter(ctx.BuildpackPath))
	}

	file = filepath.Join(ctx.BuildpackPath, "buildpack.toml")
	if _, err = toml.DecodeFile(file, &ctx.Buildpack); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode buildpack %s: %w", file, err))
		return
	}
	logger.Debug("Buildpack: %+v", ctx.Buildpack)

	ctx.Layers = Layers{config.arguments[1]}
	logger.Debug("Layers: %+v", ctx.Layers)

	ctx.Platform.Path = config.arguments[2]
	if logger.IsDebugEnabled() {
		logger.Debug("%s", PlatformFormatter(ctx.Platform))
	}

	file = filepath.Join(ctx.Platform.Path, "bindings")
	if ctx.Platform.Bindings, err = NewBindingsFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform bindings %s: %w", file, err))
		return
	}
	logger.Debug("Platform Bindings: %+v", ctx.Platform.Bindings)

	file = filepath.Join(ctx.Platform.Path, "env")
	if ctx.Platform.Environment, err = internal.NewConfigMapFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform environment %s: %w", file, err))
		return
	}
	logger.Debug("Platform Environment: %s", ctx.Platform.Environment)

	var store Store
	file = filepath.Join(ctx.Layers.Path, "store.toml")
	if _, err = toml.DecodeFile(file, &store); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode persistent metadata %s: %w", file, err))
		return
	}
	ctx.PersistentMetadata = store.Metadata
	logger.Debug("Persistent Metadata: %+v", ctx.PersistentMetadata)

	file = config.arguments[3]
	if _, err = toml.DecodeFile(file, &ctx.Plan); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode buildpack plan %s: %w", file, err))
		return
	}
	logger.Debug("Buildpack Plan: %+v", ctx.Plan)

	if ctx.StackID, ok = os.LookupEnv("CNB_STACK_ID"); !ok {
		config.exitHandler.Error(fmt.Errorf("CNB_STACK_ID not set"))
		return
	}
	logger.Debug("Stack: %s", ctx.StackID)

	result, err := f(ctx)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
	logger.Debug("Result: %+v", result)

	for _, creator := range result.Layers {
		name := creator.Name()
		layer, err := ctx.Layers.Layer(name)
		if err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to create layer %s: %w", name, err))
			return
		}

		layer, err = creator.Contribute(layer)
		if err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to invoke layer creator: %w", err))
			return
		}

		file = filepath.Join(layer.Path, "env.build")
		logger.Debug("Writing layer env.build: %s <= %+v", file, layer.BuildEnvironment)
		if err = config.environmentWriter.Write(file, layer.BuildEnvironment); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer env.build %s: %w", file, err))
			return
		}

		file = filepath.Join(layer.Path, "env.launch")
		logger.Debug("Writing layer env.launch: %s <= %+v", file, layer.LaunchEnvironment)
		if err = config.environmentWriter.Write(file, layer.LaunchEnvironment); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer env.launch %s: %w", file, err))
			return
		}

		file = filepath.Join(layer.Path, "env")
		logger.Debug("Writing layer env: %s <= %+v", file, layer.SharedEnvironment)
		if err = config.environmentWriter.Write(file, layer.SharedEnvironment); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer env %s: %w", file, err))
			return
		}

		file = filepath.Join(layer.Path, "profile.d")
		logger.Debug("Writing layer profile.d: %s <= %+v", file, layer.Profile)
		if err = config.environmentWriter.Write(file, layer.Profile); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer profile.d %s: %w", file, err))
			return
		}

		file = filepath.Join(ctx.Layers.Path, fmt.Sprintf("%s.toml", layer.Name))
		logger.Debug("Writing layer metadata: %s <= %+v", file, layer)
		if err = config.tomlWriter.Write(file, layer); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer metadata %s: %w", file, err))
			return
		}
	}

	if len(result.Processes) > 0 || len(result.Slices) > 0 {
		launch := Launch{
			Processes: result.Processes,
			Slices:    result.Slices,
		}
		file = filepath.Join(ctx.Layers.Path, "launch.toml")
		logger.Debug("Writing application metadata: %s <= %+v", file, launch)
		if err = config.tomlWriter.Write(file, launch); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write application metadata %s: %w", file, err))
			return
		}
	}

	if len(result.PersistentMetadata) > 0 {
		store = Store{
			Metadata: result.PersistentMetadata,
		}
		file = filepath.Join(ctx.Layers.Path, "store.toml")
		logger.Debug("Writing persistent metadata: %s <= %+v", file, store)
		if err = config.tomlWriter.Write(file, store); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write persistent metadata %s: %w", file, err))
			return
		}
	}

	if len(result.Plan.Entries) > 0 {
		file = config.arguments[3]
		logger.Debug("Writing buildpack plan: %s <= %+v", file, result.Plan)
		if err = config.tomlWriter.Write(file, result.Plan); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write buildpack plan %s: %w", file, err))
			return
		}
	}
}
