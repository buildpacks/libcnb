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

	"github.com/buildpacks/libcnb"
	"github.com/buildpacks/libcnb/log"
	"github.com/buildpacks/libcnb/mocks"
)

func testGenerate(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		generateFunc      libcnb.GenerateFunc
		applicationPath   string
		extensionPath     string
		outputPath        string
		buildpackPlanPath string
		extnTOMLContents  string
		commandPath       string
		environmentWriter *mocks.EnvironmentWriter
		exitHandler       *mocks.ExitHandler
		platformPath      string
		tomlWriter        *mocks.TOMLWriter
		extensionTOML     *template.Template

		workingDir string
	)

	it.Before(func() {
		generateFunc = func(libcnb.GenerateContext) (libcnb.GenerateResult, error) {
			return libcnb.NewGenerateResult(), nil
		}

		var err error
		applicationPath, err = os.MkdirTemp("", "generate-application-path")
		Expect(err).NotTo(HaveOccurred())
		applicationPath, err = filepath.EvalSymlinks(applicationPath)
		Expect(err).NotTo(HaveOccurred())

		extensionPath, err = os.MkdirTemp("", "generate-extension-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv("CNB_EXTENSION_DIR", extensionPath)).To(Succeed())

		extnTOMLContents = `
api = "{{.APIVersion}}"

[extension]
id = "test-id"
name = "test-name"
version = "1.1.1"
description = "A test buildpack"
keywords = ["test", "buildpack"]

[[extension.licenses]]
type = "Apache-2.0"
uri = "https://spdx.org/licenses/Apache-2.0.html"

[[extension.licenses]]
type = "Apache-1.1"
uri = "https://spdx.org/licenses/Apache-1.1.html"

[metadata]
test-key = "test-value"
`
		extensionTOML, err = template.New("extension.toml").Parse(extnTOMLContents)
		Expect(err).ToNot(HaveOccurred())

		var b bytes.Buffer
		err = extensionTOML.Execute(&b, map[string]string{"APIVersion": "0.8"})
		Expect(err).ToNot(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(extensionPath, "extension.toml"), b.Bytes(), 0600)).To(Succeed())

		f, err := os.CreateTemp("", "generate-buildpackplan-path")
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

		commandPath = filepath.Join("bin", "generate")

		environmentWriter = &mocks.EnvironmentWriter{}
		environmentWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		exitHandler = &mocks.ExitHandler{}
		exitHandler.On("Error", mock.Anything)

		platformPath, err = os.MkdirTemp("", "generate-platform-path")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.MkdirAll(filepath.Join(platformPath, "bindings", "alpha"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(platformPath, "bindings", "alpha", "test-secret-key"),
			[]byte("test-secret-value"), 0600)).To(Succeed())

		Expect(os.MkdirAll(filepath.Join(platformPath, "env"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(platformPath, "env", "TEST_ENV"), []byte("test-value"), 0600)).
			To(Succeed())

		tomlWriter = &mocks.TOMLWriter{}
		tomlWriter.On("Write", mock.Anything, mock.Anything).Return(nil)

		outputPath, err = os.MkdirTemp("", "generate-output-path")
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Setenv("CNB_OUTPUT_DIR", outputPath)).To(Succeed())

		Expect(os.Setenv("CNB_STACK_ID", "test-stack-id")).To(Succeed())

		Expect(os.Setenv("CNB_PLATFORM_DIR", platformPath)).To(Succeed())
		Expect(os.Setenv("CNB_BP_PLAN_PATH", buildpackPlanPath)).To(Succeed())

		workingDir, err = os.Getwd()
		Expect(err).NotTo(HaveOccurred())
		Expect(os.Chdir(applicationPath)).To(Succeed())
	})

	it.After(func() {
		Expect(os.Chdir(workingDir)).To(Succeed())
		Expect(os.Unsetenv("CNB_EXTENSION_DIR")).To(Succeed())
		Expect(os.Unsetenv("CNB_STACK_ID")).To(Succeed())
		Expect(os.Unsetenv("CNB_PLATFORM_DIR")).To(Succeed())
		Expect(os.Unsetenv("CNB_BP_PLAN_PATH")).To(Succeed())
		Expect(os.Unsetenv("CNB_OUTPUT_DIR")).To(Succeed())

		Expect(os.RemoveAll(applicationPath)).To(Succeed())
		Expect(os.RemoveAll(extensionPath)).To(Succeed())
		Expect(os.RemoveAll(buildpackPlanPath)).To(Succeed())
		Expect(os.RemoveAll(outputPath)).To(Succeed())
		Expect(os.RemoveAll(platformPath)).To(Succeed())
	})

	context("buildpack API is not within the supported range", func() {
		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(extensionPath, "extension.toml"),
				[]byte(`
api = "0.7"

[extension]
id = "test-id"
name = "test-name"
version = "1.1.1"
`),
				0600),
			).To(Succeed())
		})

		it("fails", func() {
			libcnb.Generate(generateFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath, outputPath, platformPath, buildpackPlanPath}),
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
		for _, e := range []string{"CNB_OUTPUT_DIR", "CNB_PLATFORM_DIR", "CNB_BP_PLAN_PATH"} {
			// We need to do this assignment because of the way that spec binds variables
			envVar := e
			context(fmt.Sprintf("when %s is unset", envVar), func() {
				it.Before(func() {
					Expect(os.WriteFile(filepath.Join(extensionPath, "extension.toml"),
						[]byte(`
		api = "0.8"

		[extension]
		id = "test-id"
		name = "test-name"
		version = "1.1.1"
		`),
						0600),
					).To(Succeed())
					os.Unsetenv(envVar)
				})

				it("fails", func() {
					libcnb.Generate(generateFunc,
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

		libcnb.Generate(generateFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, outputPath, platformPath, buildpackPlanPath}),
				libcnb.WithExitHandler(exitHandler),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("CNB_STACK_ID not set"))
	})

	context("has a build environment", func() {
		var ctx libcnb.GenerateContext

		it.Before(func() {
			Expect(os.WriteFile(filepath.Join(extensionPath, "extension.toml"),
				[]byte(`
	api = "0.8"
	
	[extension]
	id = "test-id"
	name = "test-name"
	version = "1.1.1"
	`),
				0600),
			).To(Succeed())

			generateFunc = func(context libcnb.GenerateContext) (libcnb.GenerateResult, error) {
				ctx = context
				return libcnb.NewGenerateResult(), nil
			}
		})

		it("creates context", func() {
			libcnb.Generate(generateFunc,
				libcnb.NewConfig(
					libcnb.WithArguments([]string{commandPath})),
			)
			Expect(ctx.ApplicationPath).To(Equal(applicationPath))
			Expect(ctx.Extension).To(Equal(libcnb.Extension{
				API: "0.8",
				Info: libcnb.ExtensionInfo{
					ID:      "test-id",
					Name:    "test-name",
					Version: "1.1.1",
				},
				Path: extensionPath,
			}))
			Expect(ctx.OutputDirectory).To(Equal(outputPath))
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

	it("fails if CNB_EXTENSION_DIR is not set", func() {
		Expect(os.Unsetenv("CNB_EXTENSION_DIR")).To(Succeed())

		libcnb.Generate(generateFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{filepath.Join(extensionPath, commandPath), outputPath, platformPath, buildpackPlanPath}),
				libcnb.WithExitHandler(exitHandler),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unable to get CNB_EXTENSION_DIR, not found"))
	})

	it("handles error from GenerateFunc", func() {
		generateFunc = func(libcnb.GenerateContext) (libcnb.GenerateResult, error) {
			return libcnb.NewGenerateResult(), errors.New("test-error")
		}

		libcnb.Generate(generateFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, outputPath, platformPath, buildpackPlanPath}),
				libcnb.WithExitHandler(exitHandler),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("test-error"))
	})

	it("writes a Dockerfile", func() {
		generateFunc = func(ctx libcnb.GenerateContext) (libcnb.GenerateResult, error) {
			os.WriteFile(filepath.Join(ctx.OutputDirectory, "build.Dockerfile"), []byte(""), 0600)
			return libcnb.NewGenerateResult(), nil
		}

		libcnb.Generate(generateFunc,
			libcnb.NewConfig(
				libcnb.WithArguments([]string{commandPath, outputPath, platformPath, buildpackPlanPath}),
				libcnb.WithTOMLWriter(tomlWriter),
				libcnb.WithLogger(log.NewDiscard())),
		)

		Expect(filepath.Join(outputPath, "build.Dockerfile")).To(BeARegularFile())
	})

}
