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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	"github.com/buildpacks/libcnb/mocks"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		applicationPath   string
		buildpackPath     string
		buildpackPlanPath string
		commandPath       string
		environmentWriter *mocks.EnvironmentWriter
		exitHandler       *mocks.ExitHandler
		layerContributor  *mocks.LayerContributor
		layersPath        string
		platformPath      string
		tomlWriter        *mocks.TOMLWriter

		workingDir string
	)

	it.Before(func() {
		var err error

		applicationPath, err = ioutil.TempDir("", "build-application-path")
		Expect(err).NotTo(HaveOccurred())
		applicationPath, err = filepath.EvalSymlinks(applicationPath)
		Expect(err).NotTo(HaveOccurred())

		buildpackPath, err = ioutil.TempDir("", "build-buildpack-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
			[]byte(`
api = "0.0.0"

[buildpack]
id = "test-id"
name = "test-name"
version = "1.1.1"
clear-env = true

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
`),
			0644),
		).To(Succeed())

		f, err := ioutil.TempFile("", "build-buildpackplan-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).NotTo(HaveOccurred())
		buildpackPlanPath = f.Name()

		Expect(ioutil.WriteFile(buildpackPlanPath,
			[]byte(`
[[entries]]
name = "test-name"
version = "test-version"

[entries.metadata]
test-key = "test-value"
`),
			0644),
		).To(Succeed())

		commandPath = filepath.Join(buildpackPath, "bin", "build")

		environmentWriter = &mocks.EnvironmentWriter{}
		environmentWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		exitHandler = &mocks.ExitHandler{}
		exitHandler.On("Error", mock.Anything)

		layerContributor = &mocks.LayerContributor{}

		layersPath, err = ioutil.TempDir("", "build-layers-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(layersPath, "store.toml"),
			[]byte(`
[metadata]
test-key = "test-value"
`),
			0644),
		).To(Succeed())

		platformPath, err = ioutil.TempDir("", "build-platform-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(platformPath, "bindings", "alpha", "metadata"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(
			filepath.Join(platformPath, "bindings", "alpha", "metadata", "test-metadata-key"),
			[]byte("test-metadata-value"),
			0644,
		)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(platformPath, "bindings", "alpha", "secret"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(
			filepath.Join(platformPath, "bindings", "alpha", "secret", "test-secret-key"),
			[]byte("test-secret-value"),
			0644,
		)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(platformPath, "env"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(platformPath, "env", "TEST_ENV"), []byte("test-value"), 0644)).
			To(Succeed())

		tomlWriter = &mocks.TOMLWriter{}
		tomlWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		Expect(os.Setenv("CNB_STACK_ID", "test-stack-id")).To(Succeed())

		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(applicationPath)).To(Succeed())
	})

	it.After(func() {
		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())

		Expect(os.RemoveAll(applicationPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPlanPath)).To(Succeed())
		Expect(os.RemoveAll(layersPath)).To(Succeed())
		Expect(os.RemoveAll(platformPath)).To(Succeed())
	})

	it("encounters the wrong number of arguments", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{}, nil
			},
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("expected 3 arguments and received 0"))
	})

	it("doesn't receive CNB_STACK_ID", func() {
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())

		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("CNB_STACK_ID not set"))
	})

	it("creates context", func() {
		var ctx libcnb.BuildContext
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				ctx = context
				return libcnb.BuildResult{}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
		)

		Expect(ctx.ApplicationPath).To(Equal(applicationPath))
		Expect(ctx.Buildpack).To(Equal(libcnb.Buildpack{
			API: "0.0.0",
			Info: libcnb.BuildpackInfo{
				ID:               "test-id",
				Name:             "test-name",
				Version:          "1.1.1",
				ClearEnvironment: true,
			},
			Orders: []libcnb.BuildpackOrder{
				{Groups: []libcnb.BuildpackOrderBuildpack{
					{
						ID:       "test-id",
						Version:  "2.2.2",
						Optional: true,
					},
				}},
			},
			Stacks: []libcnb.BuildpackStack{
				{
					ID:     "test-id",
					Mixins: []string{"test-name"},
				},
			},
			Metadata: map[string]interface{}{"test-key": "test-value"},
		}))
		Expect(ctx.BuildpackPath).To(Equal(buildpackPath))
		Expect(ctx.Layers).To(Equal(libcnb.Layers{Path: layersPath}))
		Expect(ctx.PersistentMetadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
		Expect(ctx.Plan).To(Equal(libcnb.BuildpackPlan{
			Entries: []libcnb.BuildpackPlanEntry{
				{
					Name:    "test-name",
					Version: "test-version",
					Metadata: map[string]interface{}{
						"test-key": "test-value",
					},
				},
			},
		}))
		Expect(ctx.Platform).To(Equal(libcnb.Platform{
			Bindings: libcnb.Bindings{
				"alpha": libcnb.Binding{
					Metadata: map[string]string{
						"test-metadata-key": "test-metadata-value",
					},
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

	it("handles error from BuildFunc", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{}, fmt.Errorf("test-error")
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("test-error"))
	})

	it("calls layer contributor", func() {
		layerContributor.On("Contribute", mock.Anything).Return(libcnb.Layer{}, nil)
		layerContributor.On("Name").Return("test-name")

		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{Layers: []libcnb.LayerContributor{layerContributor}}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(layerContributor.Calls).To(HaveLen(2))
	})

	it("writes env.build", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				l, err := context.Layers.Layer("test-name")
				Expect(err).NotTo(HaveOccurred())

				l.BuildEnvironment.PrependPath("test-build", "test-%s", "value")
				layerContributor.On("Contribute", mock.Anything).Return(l, nil)
				layerContributor.On("Name").Return("test-name")

				return libcnb.BuildResult{Layers: []libcnb.LayerContributor{layerContributor}}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithEnvironmentWriter(environmentWriter),
		)

		Expect(environmentWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "env.build")))
		Expect(environmentWriter.Calls[0].Arguments[1]).To(Equal(map[string]string{"test-build": "test-value"}))
	})

	it("writes env.launch", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				l, err := context.Layers.Layer("test-name")
				Expect(err).NotTo(HaveOccurred())

				l.LaunchEnvironment.PrependPath("test-launch", "test-%s", "value")
				layerContributor.On("Contribute", mock.Anything).Return(l, nil)
				layerContributor.On("Name").Return("test-name")

				return libcnb.BuildResult{Layers: []libcnb.LayerContributor{layerContributor}}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithEnvironmentWriter(environmentWriter),
		)

		Expect(environmentWriter.Calls[1].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "env.launch")))
		Expect(environmentWriter.Calls[1].Arguments[1]).To(Equal(map[string]string{"test-launch": "test-value"}))
	})

	it("writes env", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				l, err := context.Layers.Layer("test-name")
				Expect(err).NotTo(HaveOccurred())

				l.SharedEnvironment.PrependPath("test-shared", "test-%s", "value")
				layerContributor.On("Contribute", mock.Anything).Return(l, nil)
				layerContributor.On("Name").Return("test-name")

				return libcnb.BuildResult{Layers: []libcnb.LayerContributor{layerContributor}}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithEnvironmentWriter(environmentWriter),
		)

		Expect(environmentWriter.Calls[2].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "env")))
		Expect(environmentWriter.Calls[2].Arguments[1]).To(Equal(map[string]string{"test-shared": "test-value"}))
	})

	it("writes profile.d", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				l, err := context.Layers.Layer("test-name")
				Expect(err).NotTo(HaveOccurred())

				l.Profile.Add("test-profile", "test-%s", "value")
				layerContributor.On("Contribute", mock.Anything).Return(l, nil)
				layerContributor.On("Name").Return("test-name")

				return libcnb.BuildResult{Layers: []libcnb.LayerContributor{layerContributor}}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithEnvironmentWriter(environmentWriter),
		)

		Expect(environmentWriter.Calls[3].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name", "profile.d")))
		Expect(environmentWriter.Calls[3].Arguments[1]).To(Equal(map[string]string{"test-profile": "test-value"}))
	})

	it("writes layer metadata", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				l, err := context.Layers.Layer("test-name")
				Expect(err).NotTo(HaveOccurred())

				l.Build = true
				l.Cache = true
				l.Launch = true
				l.Metadata = map[string]interface{}{"test-key": "test-value"}
				layerContributor.On("Contribute", mock.Anything).Return(l, nil)
				layerContributor.On("Name").Return("test-name")

				return libcnb.BuildResult{Layers: []libcnb.LayerContributor{layerContributor}}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "test-name.toml")))

		layer, ok := tomlWriter.Calls[0].Arguments[1].(libcnb.Layer)
		Expect(ok).To(BeTrue())
		Expect(layer.Build).To(BeTrue())
		Expect(layer.Cache).To(BeTrue())
		Expect(layer.Launch).To(BeTrue())
		Expect(layer.Metadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
	})

	it("writes application metadata", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{
					Processes: []libcnb.Process{
						{
							Type:    "test-type",
							Command: "test-command",
						},
					},
					Slices: []libcnb.Slice{
						{
							Paths: []string{"test-path"},
						},
					},
				}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls[0].Arguments[0]).To(Equal(filepath.Join(layersPath, "launch.toml")))
		Expect(tomlWriter.Calls[0].Arguments[1]).To(Equal(libcnb.Launch{
			Processes: []libcnb.Process{
				{
					Type:    "test-type",
					Command: "test-command",
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

		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{PersistentMetadata: m}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls[1].Arguments[0]).To(Equal(filepath.Join(layersPath, "store.toml")))
		Expect(tomlWriter.Calls[1].Arguments[1]).To(Equal(libcnb.Store{Metadata: m}))
	})

	it("writes buildpack plan", func() {
		libcnb.Build(
			func(context libcnb.BuildContext) (libcnb.BuildResult, error) {
				return libcnb.BuildResult{
					Plan: libcnb.BuildpackPlan{
						Entries: []libcnb.BuildpackPlanEntry{
							{
								Name:     "test-name",
								Metadata: map[string]interface{}{"test-key": "test-value"},
							},
						},
					},
				}, nil
			},
			libcnb.WithArguments([]string{commandPath, layersPath, platformPath, buildpackPlanPath}),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls[2].Arguments[0]).To(Equal(buildpackPlanPath))
		Expect(tomlWriter.Calls[2].Arguments[1]).To(Equal(libcnb.BuildpackPlan{
			Entries: []libcnb.BuildpackPlanEntry{
				{
					Name:     "test-name",
					Metadata: map[string]interface{}{"test-key": "test-value"},
				},
			},
		}))
	})
}
