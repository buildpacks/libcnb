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

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		applicationPath string
		buildpackPath   string
		buildPlanPath   string
		commandPath     string
		exitHandler     *mocks.ExitHandler
		platformPath    string
		tomlWriter      *mocks.TOMLWriter

		workingDir string
	)

	it.Before(func() {
		var err error

		applicationPath, err = ioutil.TempDir("", "detect-application-path")
		Expect(err).NotTo(HaveOccurred())
		applicationPath, err = filepath.EvalSymlinks(applicationPath)
		Expect(err).NotTo(HaveOccurred())

		buildpackPath, err = ioutil.TempDir("", "detect-buildpack-path")
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

		f, err := ioutil.TempFile("", "detect-buildplan-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(f.Close()).NotTo(HaveOccurred())
		buildPlanPath = f.Name()

		commandPath = filepath.Join(buildpackPath, "bin", "detect")

		exitHandler = &mocks.ExitHandler{}
		exitHandler.On("Error", mock.Anything)
		exitHandler.On("Fail")
		exitHandler.On("Pass")

		platformPath, err = ioutil.TempDir("", "detect-platform-path")
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
		Expect(os.RemoveAll(buildPlanPath)).To(Succeed())
		Expect(os.RemoveAll(platformPath)).To(Succeed())
	})

	it("encounters the wrong number of Arguments", func() {
		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				return libcnb.DetectResult{}, nil
			},
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("expected 2 arguments and received 0"))
	})

	it("doesn't receive CNB_STACK_ID", func() {
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())

		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				return libcnb.DetectResult{}, nil
			},
			libcnb.WithArguments([]string{commandPath, platformPath, buildPlanPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("CNB_STACK_ID not set"))
	})

	it("creates context", func() {
		var ctx libcnb.DetectContext
		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				ctx = context
				return libcnb.DetectResult{}, nil
			},
			libcnb.WithArguments([]string{commandPath, platformPath, buildPlanPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(ctx.Application).To(Equal(libcnb.Application{Path: applicationPath}))
		Expect(ctx.Buildpack).To(Equal(libcnb.Buildpack{
			API: "0.0.0",
			Info: libcnb.BuildpackInfo{
				ID:               "test-id",
				Name:             "test-name",
				Version:          "1.1.1",
				ClearEnvironment: true,
			},
			Orders: []libcnb.BuildpackOrder{
				{
					Groups: []libcnb.BuildpackOrderBuildpack{
						{
							ID:       "test-id",
							Version:  "2.2.2",
							Optional: true,
						},
					},
				},
			},
			Path: buildpackPath,
			Stacks: []libcnb.BuildpackStack{
				{
					ID:     "test-id",
					Mixins: []string{"test-name"},
				},
			},
			Metadata: map[string]interface{}{"test-key": "test-value"},
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

	it("handles error from DetectFunc", func() {
		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				return libcnb.DetectResult{}, fmt.Errorf("test-error")
			},
			libcnb.WithArguments([]string{commandPath, platformPath, buildPlanPath}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("test-error"))
	})

	it("does not write empty files", func() {
		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				return libcnb.DetectResult{Pass: true}, nil
			},
			libcnb.WithArguments([]string{commandPath, platformPath, buildPlanPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls).To(HaveLen(0))
	})

	it("writes one build plan", func() {
		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				return libcnb.DetectResult{
					Pass: true,
					Plans: []libcnb.BuildPlan{
						{
							Provides: []libcnb.BuildPlanProvide{
								{Name: "test-name"},
							},
							Requires: []libcnb.BuildPlanRequire{
								{
									Name:     "test-name",
									Version:  "test-version",
									Metadata: map[string]interface{}{"test-key": "test-value"},
								},
							},
						},
					},
				}, nil
			},
			libcnb.WithArguments([]string{commandPath, platformPath, buildPlanPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls[0].Arguments.Get(0)).To(Equal(buildPlanPath))
		Expect(tomlWriter.Calls[0].Arguments.Get(1)).To(Equal(libcnb.BuildPlans{
			BuildPlan: libcnb.BuildPlan{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "test-name"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{
						Name:     "test-name",
						Version:  "test-version",
						Metadata: map[string]interface{}{"test-key": "test-value"},
					},
				},
			},
		}))
	})

	it("writes two build plans", func() {
		libcnb.Detect(
			func(context libcnb.DetectContext) (libcnb.DetectResult, error) {
				return libcnb.DetectResult{
					Pass: true,
					Plans: []libcnb.BuildPlan{
						{
							Provides: []libcnb.BuildPlanProvide{
								{Name: "test-name-1"},
							},
							Requires: []libcnb.BuildPlanRequire{
								{
									Name:     "test-name-1",
									Version:  "test-version-1",
									Metadata: map[string]interface{}{"test-key-1": "test-value-1"},
								},
							},
						},
						{
							Provides: []libcnb.BuildPlanProvide{
								{Name: "test-name-2"},
							},
							Requires: []libcnb.BuildPlanRequire{
								{
									Name:     "test-name-2",
									Version:  "test-version-2",
									Metadata: map[string]interface{}{"test-key-2": "test-value-2"},
								},
							},
						},
					},
				}, nil
			},
			libcnb.WithArguments([]string{commandPath, platformPath, buildPlanPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithTOMLWriter(tomlWriter),
		)

		Expect(tomlWriter.Calls[0].Arguments.Get(0)).To(Equal(buildPlanPath))
		Expect(tomlWriter.Calls[0].Arguments.Get(1)).To(Equal(libcnb.BuildPlans{
			BuildPlan: libcnb.BuildPlan{
				Provides: []libcnb.BuildPlanProvide{
					{Name: "test-name-1"},
				},
				Requires: []libcnb.BuildPlanRequire{
					{
						Name:     "test-name-1",
						Version:  "test-version-1",
						Metadata: map[string]interface{}{"test-key-1": "test-value-1"},
					},
				},
			},
			Or: []libcnb.BuildPlan{
				{
					Provides: []libcnb.BuildPlanProvide{
						{Name: "test-name-2"},
					},
					Requires: []libcnb.BuildPlanRequire{
						{
							Name:     "test-name-2",
							Version:  "test-version-2",
							Metadata: map[string]interface{}{"test-key-2": "test-value-2"},
						},
					},
				},
			},
		}))
	})
}
