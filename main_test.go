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

package libcnb_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"

	"github.com/buildpacks/libcnb/v2"
	"github.com/buildpacks/libcnb/v2/log"
	"github.com/buildpacks/libcnb/v2/mocks"
)

func testMain(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		applicationPath   string
		buildFunc         libcnb.BuildFunc
		buildpackPath     string
		buildpackPlanPath string
		detectFunc        libcnb.DetectFunc
		environmentWriter *mocks.EnvironmentWriter
		exitHandler       *mocks.ExitHandler
		generateFunc      libcnb.GenerateFunc
		layersPath        string
		platformPath      string
		tomlWriter        *mocks.TOMLWriter

		workingDir string
	)

	it.Before(func() {
		var err error

		applicationPath, err = os.MkdirTemp("", "main-application-path")
		Expect(err).NotTo(HaveOccurred())
		applicationPath, err = filepath.EvalSymlinks(applicationPath)
		Expect(err).NotTo(HaveOccurred())

		buildFunc = func(libcnb.BuildContext) (libcnb.BuildResult, error) {
			return libcnb.NewBuildResult(), nil
		}

		buildpackPath, err = os.MkdirTemp("", "main-buildpack-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv("CNB_BUILDPACK_DIR", buildpackPath)).To(Succeed())

		Expect(os.WriteFile(filepath.Join(buildpackPath, "buildpack.toml"),
			[]byte(`
api = "0.8"

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

[metadata]
test-key = "test-value"
`),
			0600),
		).To(Succeed())

		f, err := os.CreateTemp("", "main-buildpackplan-path")
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

		detectFunc = func(libcnb.DetectContext) (libcnb.DetectResult, error) {
			return libcnb.DetectResult{}, nil
		}

		generateFunc = func(libcnb.GenerateContext) (libcnb.GenerateResult, error) {
			return libcnb.GenerateResult{}, nil
		}

		environmentWriter = &mocks.EnvironmentWriter{}
		environmentWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		exitHandler = &mocks.ExitHandler{}
		exitHandler.On("Error", mock.Anything)
		exitHandler.On("Pass", mock.Anything)
		exitHandler.On("Fail", mock.Anything)

		layersPath, err = os.MkdirTemp("", "main-layers-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(layersPath, "store.toml"),
			[]byte(`
[metadata]
test-key = "test-value"
`),
			0600),
		).To(Succeed())

		platformPath, err = os.MkdirTemp("", "main-platform-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(platformPath, "bindings", "alpha", "metadata"), 0755)).To(Succeed())
		Expect(os.WriteFile(
			filepath.Join(platformPath, "bindings", "alpha", "metadata", "test-metadata-key"),
			[]byte("test-metadata-value"),
			0600,
		)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(platformPath, "bindings", "alpha", "secret"), 0755)).To(Succeed())
		Expect(os.WriteFile(
			filepath.Join(platformPath, "bindings", "alpha", "secret", "test-secret-key"),
			[]byte("test-secret-value"),
			0600,
		)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(platformPath, "env"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(platformPath, "env", "TEST_ENV"), []byte("test-value"), 0600)).
			To(Succeed())

		tomlWriter = &mocks.TOMLWriter{}
		tomlWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		Expect(os.Setenv("CNB_STACK_ID", "test-stack-id")).To(Succeed())
		Expect(os.Setenv("CNB_LAYERS_DIR", layersPath)).To(Succeed())
		Expect(os.Setenv("CNB_PLATFORM_DIR", platformPath)).To(Succeed())
		Expect(os.Setenv("CNB_BP_PLAN_PATH", buildpackPlanPath)).To(Succeed())
		Expect(os.Setenv("CNB_BUILD_PLAN_PATH", buildpackPlanPath)).To(Succeed())

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
		Expect(os.Unsetenv("CNB_BUILD_PLAN_PATH")).To(Succeed())

		Expect(os.RemoveAll(applicationPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPlanPath)).To(Succeed())
		Expect(os.RemoveAll(layersPath)).To(Succeed())
		Expect(os.RemoveAll(platformPath)).To(Succeed())
	})

	it("encounters the wrong number of arguments", func() {
		libcnb.BuildpackMain(detectFunc, buildFunc,
			libcnb.WithArguments([]string{}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("expected command name"))
	})

	it("calls builder for build command", func() {
		commandPath := filepath.Join("bin", "build")

		libcnb.BuildpackMain(detectFunc, buildFunc,
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)

		Expect(exitHandler.Calls).To(BeEmpty())
	})

	it("calls detector for detect command", func() {
		detectFunc = func(libcnb.DetectContext) (libcnb.DetectResult, error) {
			return libcnb.DetectResult{Pass: true}, nil
		}
		commandPath := filepath.Join("bin", "detect")

		libcnb.BuildpackMain(detectFunc, buildFunc,
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)
	})

	it("calls generator for generate command", func() {
		generateFunc = func(libcnb.GenerateContext) (libcnb.GenerateResult, error) {
			return libcnb.GenerateResult{}, nil
		}
		commandPath := filepath.Join("bin", "generate")

		libcnb.ExtensionMain(nil, generateFunc,
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)
	})

	it("calls exitHandler.Pass() on detection pass", func() {
		detectFunc = func(libcnb.DetectContext) (libcnb.DetectResult, error) {
			return libcnb.DetectResult{Pass: true}, nil
		}
		commandPath := filepath.Join("bin", "detect")

		libcnb.BuildpackMain(detectFunc, buildFunc,
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)

		Expect(exitHandler.Calls[0].Method).To(BeIdenticalTo("Pass"))
	})

	it("calls exitHandler.Fail() on detection fail", func() {
		detectFunc = func(libcnb.DetectContext) (libcnb.DetectResult, error) {
			return libcnb.DetectResult{Pass: false}, nil
		}
		commandPath := filepath.Join("bin", "detect")

		libcnb.BuildpackMain(detectFunc, buildFunc,
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)

		Expect(exitHandler.Calls[0].Method).To(BeIdenticalTo("Fail"))
	})

	it("encounters an unknown command", func() {
		commandPath := filepath.Join("bin", "test-command")

		libcnb.BuildpackMain(detectFunc, buildFunc,
			libcnb.WithArguments([]string{commandPath}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithLogger(log.NewDiscard()),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unsupported command test-command"))
	})
}
