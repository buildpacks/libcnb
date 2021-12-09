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
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"

	"github.com/buildpacks/libcnb/internal"
	"github.com/buildpacks/libcnb/log"
)

// BuildContext contains the inputs to build.
type BuildContext struct {
	// WorkingDir is the location of the application source code as provided by
	// the lifecycle.
	WorkingDir string

	// Buildpack is metadata about the buildpack, from buildpack.toml.
	Buildpack Buildpack

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
	// Labels are the image labels contributed by the buildpack.
	Labels []Label

	// Layers is the collection of LayerCreators contributed by the buildpack.
	Layers []Layer

	// PersistentMetadata is metadata that is persisted even across cache cleaning.
	PersistentMetadata map[string]interface{}

	// Processes are the process types contributed by the buildpack.
	Processes []Process

	// Slices are the application slices contributed by the buildpack.
	Slices []Slice

	// Unmet contains buildpack plan entries that were not satisfied by the buildpack and therefore should be
	// passed to subsequent providers.
	Unmet []UnmetPlanEntry
}

// NewBuildResult creates a new BuildResult instance, initializing empty fields.
func NewBuildResult() BuildResult {
	return BuildResult{
		PersistentMetadata: make(map[string]interface{}),
	}
}

func (b BuildResult) String() string {
	var l []string
	for _, c := range b.Layers {
		l = append(l, reflect.TypeOf(c).Name())
	}

	return fmt.Sprintf(
		"{Labels:%+v Layers:%s PersistentMetadata:%+v Processes:%+v Slices:%+v, Unmet:%+v}",
		b.Labels, l, b.PersistentMetadata, b.PersistentMetadata, b.Slices, b.Unmet,
	)
}

// BuildFunc takes a context and returns a result, performing buildpack build behaviors.
type BuildFunc func(context BuildContext) (BuildResult, error)

// Build is called by the main function of a buildpack, for build.
func Build(build BuildFunc, options ...Option) {
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
	logger := log.New(os.Stdout)

	ctx.WorkingDir, err = os.Getwd()
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to get working directory\n%w", err))
		return
	}
	if logger.IsDebugEnabled() {
		logger.Debug(ApplicationPathFormatter(ctx.WorkingDir))
	}

	if s, ok := os.LookupEnv("CNB_BUILDPACK_DIR"); ok {
		ctx.Buildpack.Path = filepath.Clean(s)
	} else { // TODO: Remove branch once lifecycle has been updated to support this
		ctx.Buildpack.Path = filepath.Clean(strings.TrimSuffix(config.arguments[0], filepath.Join("bin", "build")))
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
		config.exitHandler.Error(errors.New("this version of libcnb is only compatible with buildpack APIs 0.5 and 0.6"))
		return
	}

	ctx.Layers = Layers{config.arguments[1]}
	logger.Debugf("Layers: %+v", ctx.Layers)

	ctx.Platform.Path = config.arguments[2]
	if logger.IsDebugEnabled() {
		logger.Debug(PlatformFormatter(ctx.Platform))
	}

	if ctx.Platform.Bindings, err = NewBindingsForBuild(ctx.Platform.Path); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform bindings %s\n%w", ctx.Platform.Path, err))
		return
	}
	logger.Debugf("Platform Bindings: %+v", ctx.Platform.Bindings)

	file = filepath.Join(ctx.Platform.Path, "env")
	if ctx.Platform.Environment, err = internal.NewConfigMapFromPath(file); err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to read platform environment %s\n%w", file, err))
		return
	}
	logger.Debugf("Platform Environment: %s", ctx.Platform.Environment)

	var store Store
	file = filepath.Join(ctx.Layers.Path, "store.toml")
	if _, err = toml.DecodeFile(file, &store); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode persistent metadata %s\n%w", file, err))
		return
	}
	ctx.PersistentMetadata = store.Metadata
	logger.Debugf("Persistent Metadata: %+v", ctx.PersistentMetadata)

	file = config.arguments[3]
	if _, err = toml.DecodeFile(file, &ctx.Plan); err != nil && !os.IsNotExist(err) {
		config.exitHandler.Error(fmt.Errorf("unable to decode buildpack plan %s\n%w", file, err))
		return
	}
	logger.Debugf("Buildpack Plan: %+v", ctx.Plan)

	if ctx.StackID, ok = os.LookupEnv("CNB_STACK_ID"); !ok {
		config.exitHandler.Error(fmt.Errorf("CNB_STACK_ID not set"))
		return
	}
	logger.Debugf("Stack: %s", ctx.StackID)

	result, err := build(ctx)
	if err != nil {
		config.exitHandler.Error(err)
		return
	}
	logger.Debugf("Result: %+v", result)

	file = filepath.Join(ctx.Layers.Path, "*.toml")
	existing, err := filepath.Glob(file)
	if err != nil {
		config.exitHandler.Error(fmt.Errorf("unable to list files in %s\n%w", file, err))
		return
	}
	var contributed []string

	for _, layer := range result.Layers {
		file = filepath.Join(layer.Path, "env.build")
		logger.Debugf("Writing layer env.build: %s <= %+v", file, layer.BuildEnvironment)
		if err = config.environmentWriter.Write(file, layer.BuildEnvironment); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer env.build %s\n%w", file, err))
			return
		}

		file = filepath.Join(layer.Path, "env.launch")
		logger.Debugf("Writing layer env.launch: %s <= %+v", file, layer.LaunchEnvironment)
		if err = config.environmentWriter.Write(file, layer.LaunchEnvironment); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer env.launch %s\n%w", file, err))
			return
		}

		file = filepath.Join(layer.Path, "env")
		logger.Debugf("Writing layer env: %s <= %+v", file, layer.SharedEnvironment)
		if err = config.environmentWriter.Write(file, layer.SharedEnvironment); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer env %s\n%w", file, err))
			return
		}

		file = filepath.Join(layer.Path, "profile.d")
		logger.Debugf("Writing layer profile.d: %s <= %+v", file, layer.Profile)
		if err = config.environmentWriter.Write(file, layer.Profile); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer profile.d %s\n%w", file, err))
			return
		}

		file = filepath.Join(ctx.Layers.Path, fmt.Sprintf("%s.toml", layer.Name))
		logger.Debugf("Writing layer metadata: %s <= %+v", file, layer)
		var toWrite interface{} = layer
		if API == "0.5" {
			toWrite = internal.LayerAPI5{
				Build:    layer.LayerTypes.Build,
				Cache:    layer.LayerTypes.Cache,
				Launch:   layer.LayerTypes.Launch,
				Metadata: layer.Metadata,
			}
		}
		if err = config.tomlWriter.Write(file, toWrite); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write layer metadata %s\n%w", file, err))
			return
		}
		contributed = append(contributed, file)
	}

	for _, e := range existing {
		if strings.HasSuffix(e, "store.toml") || contains(contributed, e) {
			continue
		}

		logger.Debugf("Removing %s", e)

		if err := os.RemoveAll(e); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to remove %s\n%w", e, err))
			return
		}
	}

	launch := LaunchTOML{
		Labels:    result.Labels,
		Processes: result.Processes,
		Slices:    result.Slices,
	}

	if !launch.isEmpty() {
		file = filepath.Join(ctx.Layers.Path, "launch.toml")
		logger.Debugf("Writing application metadata: %s <= %+v", file, launch)

		if API == "0.5" {
			for _, process := range launch.Processes {
				if process.Default {
					logger.Info("WARNING: Launch layer is setting default=true, but that is not supported until API version 0.6. This setting will be ignored.")
				}
			}
		}

		if err = config.tomlWriter.Write(file, launch); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write application metadata %s\n%w", file, err))
			return
		}
	}

	buildTOML := BuildTOML{
		Unmet: result.Unmet,
	}

	if !buildTOML.isEmpty() {
		file = filepath.Join(ctx.Layers.Path, "build.toml")
		logger.Debugf("Writing build metadata: %s <= %+v", file, build)
		if err = config.tomlWriter.Write(file, buildTOML); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write build metadata %s\n%w", file, err))
			return
		}
	}

	if len(result.PersistentMetadata) > 0 {
		store = Store{
			Metadata: result.PersistentMetadata,
		}
		file = filepath.Join(ctx.Layers.Path, "store.toml")
		logger.Debugf("Writing persistent metadata: %s <= %+v", file, store)
		if err = config.tomlWriter.Write(file, store); err != nil {
			config.exitHandler.Error(fmt.Errorf("unable to write persistent metadata %s\n%w", file, err))
			return
		}
	}
}

func contains(candidates []string, s string) bool {
	for _, c := range candidates {
		if s == c {
			return true
		}
	}

	return false
}
