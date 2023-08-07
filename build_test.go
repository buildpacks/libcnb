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

package libcnb_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"

	"github.com/buildpacks/libcnb/v2"
	"github.com/buildpacks/libcnb/v2/log"
	"github.com/buildpacks/libcnb/v2/mocks"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buildFunc         libcnb.BuildFunc
		applicationPath   string
		buildpackPath     string
		buildpackPlanPath string
		bpTOMLContents    string
		commandPath       string
		environmentWriter *mocks.EnvironmentWriter
		exitHandler       *mocks.ExitHandler
		layersPath        string
		platformPath      string
		tomlWriter        *mocks.TOMLWriter
		buildpackTOML     *template.Template

		workingDir string
	)

	it.Before(func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.NewBuildResult(), nil
		}

		var err error
		applicationPath, err = os.MkdirTemp("", "build-application-path")
		Expect(err).NotTo(HaveOccurred())
		applicationPath, err = filepath.EvalSymlinks(applicationPath)
		Expect(err).NotTo(HaveOccurred())

		buildpackPath, err = os.MkdirTemp("", "build-buildpack-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv("CNB_BUILDPACK_DIR", buildpackPath)).To(Succeed())

		bpTOMLContents = `
api = "{{.APIVersion}}"

[buildpack]
id = "test-id"
name = "test-name"
version = "1.1.1"
clear-env = true
description = "A test buildpack"
keywords = ["test", "buildpack"]

[[buildpack.licenses]]
type = "Apache-2.0"
uri = "https://spdx.org/licenses/Apache-2.0.html"

[[buildpack.licenses]]
type = "Apache-1.1"
uri = "https://spdx.org/licenses/Apache-1.1.html"

[[order]]
[[order.group]]
id = "test-id"
version = "2.2.2"
optional = true

[[stacks]]
id = "test-id"
mixins = ["test-name"]

[metadata]
test-key = "test-value"
`
		buildpackTOML, err = template.New("buildpack.toml").Parse(bpTOMLContents)
		Expect(err).ToNot(HaveOccurred())

		var b bytes.Buffer
		err = buildpackTOML.Execute(&b, map[string]string{"APIVersion": "0.8"})
		Expect(err).ToNot(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"), b.Bytes(), 0600)).To(Succeed())

		f, err := os.CreateTemp("", "build-buildpackplan-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).NotTo(HaveOccurred())
		buildpackPlanPath = f.Name()

		Expect(os.WriteFile(buildpackPlanPath,
			[]byte(`
[[entries]]
name = "test-name"
version = "test-version"

[entries.metadata]
test-key = "test-value"
`),
			0600),
		).To(Succeed())

		commandPath = filepath.Join("bin", "build")

		environmentWriter = &mocks.EnvironmentWriter{}
		environmentWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		exitHandler = &mocks.ExitHandler{}
		exitHandler.On("Error", mock.Anything)

		layersPath, err = os.MkdirTemp("", "build-layers-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(layersPath, "store.toml"),
			[]byte(`
[metadata]
test-key = "test-value"
`),
			0600),
		).To(Succeed())

		platformPath, err = os.MkdirTemp("", "build-platform-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(platformPath, "bindings", "alpha"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(platformPath, "bindings", "alpha", "test-secret-key"),
			[]byte("test-secret-value"), 0600)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(platformPath, "env"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(platformPath, "env", "TEST_ENV"), []byte("test-value"), 0600)).
			To(Succeed())

		tomlWriter = &mocks.TOMLWriter{}
		tomlWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		Expect(os.Setenv("CNB_STACK_ID", "test-stack-id")).To(Succeed())
		Expect(os.Setenv("CNB_LAYERS_DIR", layersPath)).To(Succeed())
		Expect(os.Setenv("CNB_PLATFORM_DIR", platformPath)).To(Succeed())
		Expect(os.Setenv("CNB_BP_PLAN_PATH", buildpackPlanPath)).To(Succeed())

		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(applicationPath)).To(Succeed())
	})

	it.After(func() {
		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.Unsetenv("CNB_BUILDPACK_DIR")).To(Succeed())
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())
		Expect(os.Unsetenv("CNB_PLATFORM_DIR")).To(Succeed())
		Expect(os.Unsetenv("CNB_BP_PLAN_PATH")).To(Succeed())
		Expect(os.Unsetenv("CNB_LAYERS_DIR")).To(Succeed())

		Expect(os.RemoveAll(applicationPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPlanPath)).To(Succeed())
		Expect(os.RemoveAll(layersPath)).To(Succeed())
		Expect(os.RemoveAll(platformPath)).To(Succeed())
	})

	context("buildpack API is not within the supported range", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
				[]byte(`
api = "0.7"

[buildpack]
id = "test-id"
name = "test-name"
version = "1.1.1"
`),
				0600),
			).To(Succeed())
		})

		it("fails", func() {
			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
					libcnb.WithExitHandler(exitHandler),
					libcnb.WithLogger(log.NewDiscard())),
			)

			if libcnb.MinSupportedBPVersion == libcnb.MaxSupportedBPVersion {
				Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError(
					fmt.Sprintf("this version of libcnb is only compatible with buildpack API == %s", libcnb.MinSupportedBPVersion)))
			} else {
				Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError(
					fmt.Sprintf("this version of libcnb is only compatible with buildpack APIs >= %s, <= %s", libcnb.MinSupportedBPVersion, libcnb.MaxSupportedBPVersion),
				))
			}
		})
	})

	context("errors if required env vars are not set", func() {
		for _, e := range []string{"CNB_LAYERS_DIR", "CNB_PLATFORM_DIR", "CNB_BP_PLAN_PATH"} {
			// We need to do this assignment because of the way that spec binds variables
			envVar := e
			context(fmt.Sprintf("when %s is unset", envVar), func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
						[]byte(`
		api = "0.8"
		
		[buildpack]
		id = "test-id"
		name = "test-name"
		version = "1.1.1"
		`),
						0600),
					).To(Succeed())
					os.Unsetenv(envVar)
				})

				it("fails", func() {
					libcnb.Build(buildFunc,
						libcnb.NewConfig(
							libcnb.WithArguments([]string{commandPath}),
							libcnb.WithExitHandler(exitHandler)),
					)
					Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError(
						fmt.Sprintf("expected %s to be set", envVar),
					))
				})
			})
		}
	})

	it("doesn't receive CNB_STACK_ID", func() {
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithExitHandler(exitHandler),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("CNB_STACK_ID not set"))
	})

	context("has a build environment", func() {
		var ctx libcnb.BuildContext

		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
				[]byte(`
	api = "0.8"
	
	[buildpack]
	id = "test-id"
	name = "test-name"
	version = "1.1.1"
	`),
				0600),
			).To(Succeed())

			buildFunc = func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				ctx = context
				return libcnb.NewBuildResult(), nil
			}
		})

		it("creates context", func() {
			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath})),
			)

			Expect(ctx.ApplicationPath).To(Equal(applicationPath))
			Expect(ctx.Buildpack).To(Equal(libcnb.Buildpack{
				API: "0.8",
				Info: libcnb.BuildpackInfo{
					ID:      "test-id",
					Name:    "test-name",
					Version: "1.1.1",
				},
				Path: buildpackPath,
			}))
			Expect(ctx.Layers).To(Equal(libcnb.Layers{Path: layersPath}))
			Expect(ctx.PersistentMetadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
			Expect(ctx.Plan).To(Equal(libcnb.BuildpackPlan{
				Entries: []libcnb.BuildpackPlanEntry{
					{
						Name: "test-name",
						Metadata: map[string]interface{}{
							"test-key": "test-value",
						},
					},
				},
			}))
			Expect(ctx.Platform).To(Equal(libcnb.Platform{
				Bindings: libcnb.Bindings{
					libcnb.Binding{
						Name: "alpha",
						Path: filepath.Join(platformPath, "bindings", "alpha"),
						Secret: map[string]string{
							"test-secret-key": "test-secret-value",
						},
					},
				},
				Environment: map[string]string{"TEST_ENV": "test-value"},
				Path:        platformPath,
			}))
			Expect(ctx.StackID).To(Equal("test-stack-id"))
		})
	})

	it("fails if CNB_BUILDPACK_DIR is not set", func() {
		Expect(os.Unsetenv("CNB_BUILDPACK_DIR")).To(Succeed())

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{filepath.Join(buildpackPath, commandPath), layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithExitHandler(exitHandler),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unable to get CNB_BUILDPACK_DIR, not found"))
	})

	it("handles error from BuildFunc", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.NewBuildResult(), errors.New("test-error")
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithExitHandler(exitHandler),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("test-error"))
	})

	it("writes env.build", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			layer := libcnb.Layer{Path: filepath.Join(layersPath, "test-name"), BuildEnvironment: libcnb.Environment{}}
			layer.BuildEnvironment.Defaultf("test-build", "test-%s", "value")
			return libcnb.BuildResult{Layers: []libcnb.Layer{layer}}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithEnvironmentWriter(environmentWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(environmentWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "env.build")))
		Expect(environmentWriter.Calls[0].Arguments[1]).To(Equal(map[string]string{"test-build.default": "test-value"}))
	})

	it("writes env.launch", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			layer := libcnb.Layer{Path: filepath.Join(layersPath, "test-name"), LaunchEnvironment: libcnb.Environment{}}
			layer.LaunchEnvironment.Defaultf("test-launch", "test-%s", "value")
			return libcnb.BuildResult{Layers: []libcnb.Layer{layer}}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithEnvironmentWriter(environmentWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(environmentWriter.Calls[1].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "env.launch")))
		Expect(environmentWriter.Calls[1].Arguments[1]).To(Equal(map[string]string{"test-launch.default": "test-value"}))
	})

	it("writes env", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			layer := libcnb.Layer{Path: filepath.Join(layersPath, "test-name"), SharedEnvironment: libcnb.Environment{}}
			layer.SharedEnvironment.Defaultf("test-shared", "test-%s", "value")
			return libcnb.BuildResult{Layers: []libcnb.Layer{layer}}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithEnvironmentWriter(environmentWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(environmentWriter.Calls[2].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "env")))
		Expect(environmentWriter.Calls[2].Arguments[1]).To(Equal(map[string]string{"test-shared.default": "test-value"}))
	})

	it("writes profile.d", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			layer := libcnb.Layer{Path: filepath.Join(layersPath, "test-name"), Profile: libcnb.Profile{}}
			layer.Profile.Addf("test-profile", "test-%s", "value")
			return libcnb.BuildResult{Layers: []libcnb.Layer{layer}}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithEnvironmentWriter(environmentWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(environmentWriter.Calls[3].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "profile.d")))
		Expect(environmentWriter.Calls[3].Arguments[1]).To(Equal(map[string]string{"test-profile": "test-value"}))
	})

	it("writes layer metadata", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			layer := libcnb.Layer{
				Name: "test-name",
				Path: filepath.Join(layersPath, "test-name"),
				LayerTypes: libcnb.LayerTypes{
					Build:  true,
					Cache:  true,
					Launch: true,
				},
				Metadata: map[string]interface{}{"test-key": "test-value"},
			}
			return libcnb.BuildResult{Layers: []libcnb.Layer{layer}}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name.toml")))

		layer, ok := tomlWriter.Calls[0].Arguments[1].(libcnb.Layer)
		Expect(ok).To(BeTrue())
		Expect(layer.LayerTypes.Build).To(BeTrue())
		Expect(layer.LayerTypes.Cache).To(BeTrue())
		Expect(layer.LayerTypes.Launch).To(BeTrue())
		Expect(layer.Metadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
	})

	it("writes launch.toml with working-directory setting", func() {
		var b bytes.Buffer
		err := buildpackTOML.Execute(&b, map[string]string{"APIVersion": "0.8"})
		Expect(err).ToNot(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"), b.Bytes(), 0600)).To(Succeed())

		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.BuildResult{
				Layers: []libcnb.Layer{},
				Processes: []libcnb.Process{
					{
						Type:             "test-type",
						Command:          []string{"test-command-in-dir"},
						Default:          true,
						WorkingDirectory: "/my/directory/",
					},
				},
			}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter)),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "launch.toml")))
		Expect(tomlWriter.Calls[0].Arguments[1]).To(Equal(libcnb.LaunchTOML{
			Processes: []libcnb.Process{
				{
					Type:             "test-type",
					Command:          []string{"test-command-in-dir"},
					Default:          true,
					WorkingDirectory: "/my/directory/",
				},
			},
		}))
	})

	it("writes launch.toml", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.BuildResult{
				Labels: []libcnb.Label{
					{
						Key:   "test-key",
						Value: "test-value",
					},
				},
				Processes: []libcnb.Process{
					{
						Type:    "test-type",
						Command: []string{"test-command"},
						Default: true,
					},
				},
				Slices: []libcnb.Slice{
					{
						Paths: []string{"test-path"},
					},
				},
			}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "launch.toml")))
		Expect(tomlWriter.Calls[0].Arguments[1]).To(Equal(libcnb.LaunchTOML{
			Labels: []libcnb.Label{
				{
					Key:   "test-key",
					Value: "test-value",
				},
			},
			Processes: []libcnb.Process{
				{
					Type:    "test-type",
					Command: []string{"test-command"},
					Default: true,
				},
			},
			Slices: []libcnb.Slice{
				{
					Paths: []string{"test-path"},
				},
			},
		}))
	})

	it("writes persistent metadata", func() {
		m := map[string]interface{}{"test-key": "test-value"}

		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.BuildResult{PersistentMetadata: m}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "store.toml")))
		Expect(tomlWriter.Calls[0].Arguments[1]).To(Equal(libcnb.Store{Metadata: m}))
	})

	it("does not write empty files", func() {
		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(tomlWriter.Calls).To(HaveLen(0))
	})

	it("removes stale layers", func() {
		Expect(os.WriteFile(filepath.Join(layersPath, "alpha.toml"), []byte(""), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(layersPath, "bravo.toml"), []byte(""), 0600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(layersPath, "store.toml"), []byte(""), 0600)).To(Succeed())

		layer := libcnb.Layer{Name: "alpha"}

		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.BuildResult{Layers: []libcnb.Layer{layer}}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(tomlWriter.Calls).To(HaveLen(1))
		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "alpha.toml")))
		Expect(filepath.Join(layersPath, "bravo.toml")).NotTo(BeARegularFile())
		Expect(filepath.Join(layersPath, "store.toml")).To(BeARegularFile())
	})

	it("writes build.toml", func() {
		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.BuildResult{
				Unmet: []libcnb.UnmetPlanEntry{
					{
						Name: "test-entry",
					},
				},
			}, nil
		}

		libcnb.Build(buildFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "build.toml")))
		Expect(tomlWriter.Calls[0].Arguments[1]).To(Equal(libcnb.BuildTOML{
			Unmet: []libcnb.UnmetPlanEntry{
				{
					Name: "test-entry",
				},
			},
		}))
	})

	context("Validates SBOM entries", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
				[]byte(`
api = "0.8"

[buildpack]
id = "test-id"
name = "test-name"
version = "1.1.1"
sbom-formats = ["application/vnd.cyclonedx+json"]
`),
				0600),
			).To(Succeed())

			buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{}, nil
			}
		})

		it("has SBOM files", func() {
			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
					libcnb.WithExitHandler(exitHandler),
					libcnb.WithLogger(log.NewDiscard())),
			)

			Expect(exitHandler.Calls).To(BeEmpty())
		})

		it("has no accepted formats", func() {
			Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
				[]byte(`
api = "0.8"

[buildpack]
id = "test-id"
name = "test-name"
version = "1.1.1"
sbom-formats = []
`),
				0600),
			).To(Succeed())

			Expect(os.WriteFile(filepath.Join(layersPath, "launch.sbom.spdx.json"), []byte{}, 0600)).To(Succeed())

			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
					libcnb.WithExitHandler(exitHandler),
					libcnb.WithLogger(log.NewDiscard())),
			)

			Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unable to validate SBOM\nunable to find actual SBOM Type application/spdx+json in list of supported SBOM types []"))
		})

		it("has no matching formats", func() {
			Expect(os.WriteFile(filepath.Join(layersPath, "launch.sbom.spdx.json"), []byte{}, 0600)).To(Succeed())

			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
					libcnb.WithExitHandler(exitHandler),
					libcnb.WithLogger(log.NewDiscard())),
			)

			Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unable to validate SBOM\nunable to find actual SBOM Type application/spdx+json in list of supported SBOM types [application/vnd.cyclonedx+json]"))
		})

		it("has a matching format", func() {
			Expect(os.WriteFile(filepath.Join(layersPath, "launch.sbom.cdx.json"), []byte{}, 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(layersPath, "layer.sbom.cdx.json"), []byte{}, 0600)).To(Succeed())
			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
					libcnb.WithExitHandler(exitHandler),
					libcnb.WithLogger(log.NewDiscard())),
			)

			Expect(exitHandler.Calls).To(BeEmpty())
		})

		it("has a junk format", func() {
			Expect(os.WriteFile(filepath.Join(layersPath, "launch.sbom.random.json"), []byte{}, 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(layersPath, "layer.sbom.cdx.json"), []byte{}, 0600)).To(Succeed())
			libcnb.Build(buildFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
					libcnb.WithExitHandler(exitHandler),
					libcnb.WithLogger(log.NewDiscard())),
			)

			Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unable to validate SBOM\nunable to parse SBOM unknown\nunable to translate from random.json to SBOMFormat"))
		})
	})
}
