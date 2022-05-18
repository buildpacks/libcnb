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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb"
)

func testLayer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layers libcnb.Layers
		path   string
	)

	context("Exec", func() {
		var exec libcnb.Exec

		it.Before(func() {
			exec = libcnb.Exec{Path: "test-path"}
		})

		it("returns filename", func() {
			Expect(exec.FilePath("test-name")).To(Equal(filepath.Join("test-path", "test-name")))
		})

		it("returns process-specific filename", func() {
			Expect(exec.ProcessFilePath("test-process", "test-name")).
				To(Equal(filepath.Join("test-path", "test-process", "test-name")))
		})
	})

	context("Profile", func() {
		var profile libcnb.Profile

		it.Before(func() {
			profile = libcnb.Profile{}
		})

		it("adds content", func() {
			profile.Add("test-name", "test-value")
			Expect(profile).To(Equal(libcnb.Profile{"test-name": "test-value"}))
		})

		it("adds formatted content", func() {
			profile.Addf("test-name", "test-%s", "value")
			Expect(profile).To(Equal(libcnb.Profile{"test-name": "test-value"}))
		})

		it("adds process-specific content", func() {
			profile.ProcessAdd("test-process", "test-name", "test-value")
			Expect(profile).To(Equal(libcnb.Profile{filepath.Join("test-process", "test-name"): "test-value"}))
		})

		it("adds process-specific formatted content", func() {
			profile.ProcessAddf("test-process", "test-name", "test-%s", "value")
			Expect(profile).To(Equal(libcnb.Profile{filepath.Join("test-process", "test-name"): "test-value"}))
		})
	})

	context("Reset", func() {
		var layer libcnb.Layer

		it.Before(func() {
			var err error
			path, err = ioutil.TempDir("", "layers")
			Expect(err).NotTo(HaveOccurred())

			layers = libcnb.Layers{Path: path}
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		context("when there is no previous build", func() {
			it.Before(func() {
				layer = libcnb.Layer{
					Name: "test-name",
					Path: filepath.Join(layers.Path, "test-name"),
					LayerTypes: libcnb.LayerTypes{
						Launch: true,
						Build:  true,
						Cache:  true,
					},
				}
			})

			it("initializes an empty layer", func() {
				var err error
				layer, err = layer.Reset()
				Expect(err).NotTo(HaveOccurred())

				Expect(layer).To(Equal(libcnb.Layer{
					Name: "test-name",
					Path: filepath.Join(layers.Path, "test-name"),
					LayerTypes: libcnb.LayerTypes{
						Launch: false,
						Build:  false,
						Cache:  false,
					},
					SharedEnvironment: libcnb.Environment{},
					BuildEnvironment:  libcnb.Environment{},
					LaunchEnvironment: libcnb.Environment{},
				}))

				Expect(filepath.Join(layers.Path, "test-name")).To(BeADirectory())
			})
		})

		context("when cache is retrieved from previous build", func() {
			it.Before(func() {
				sharedEnvDir := filepath.Join(layers.Path, "test-name", "env")
				Expect(os.MkdirAll(sharedEnvDir, os.ModePerm)).To(Succeed())

				err := os.WriteFile(filepath.Join(sharedEnvDir, "OVERRIDE_VAR.override"), []byte("override-value"), 0600)
				Expect(err).NotTo(HaveOccurred())

				buildEnvDir := filepath.Join(layers.Path, "test-name", "env.build")
				Expect(os.MkdirAll(buildEnvDir, os.ModePerm)).To(Succeed())

				err = os.WriteFile(filepath.Join(buildEnvDir, "DEFAULT_VAR.default"), []byte("default-value"), 0600)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(buildEnvDir, "INVALID_VAR.invalid"), []byte("invalid-value"), 0600)
				Expect(err).NotTo(HaveOccurred())

				launchEnvDir := filepath.Join(layers.Path, "test-name", "env.launch")
				Expect(os.MkdirAll(launchEnvDir, os.ModePerm)).To(Succeed())

				err = os.WriteFile(filepath.Join(launchEnvDir, "APPEND_VAR.append"), []byte("append-value"), 0600)
				Expect(err).NotTo(HaveOccurred())

				err = os.WriteFile(filepath.Join(launchEnvDir, "APPEND_VAR.delim"), []byte("!"), 0600)
				Expect(err).NotTo(HaveOccurred())

				layer = libcnb.Layer{
					Name: "test-name",
					Path: filepath.Join(layers.Path, "test-name"),
					LayerTypes: libcnb.LayerTypes{
						Launch: true,
						Build:  true,
						Cache:  true,
					},
					SharedEnvironment: libcnb.Environment{
						"OVERRIDE_VAR.override": "override-value",
					},
					BuildEnvironment: libcnb.Environment{
						"DEFAULT_VAR.default": "default-value",
					},
					LaunchEnvironment: libcnb.Environment{
						"APPEND_VAR.append": "append-value",
						"APPEND_VAR.delim":  "!",
					},
					Metadata: map[string]interface{}{
						"some-key": "some-value",
					},
				}
			})

			context("when Reset is called on a layer", func() {
				it("resets all of the layer data and clears the directory", func() {
					layer, err := layer.Reset()
					Expect(err).NotTo(HaveOccurred())

					Expect(layer).To(Equal(libcnb.Layer{
						Name: "test-name",
						Path: filepath.Join(layers.Path, "test-name"),
						LayerTypes: libcnb.LayerTypes{
							Launch: false,
							Build:  false,
							Cache:  false,
						},
						SharedEnvironment: libcnb.Environment{},
						BuildEnvironment:  libcnb.Environment{},
						LaunchEnvironment: libcnb.Environment{},
					}))

					Expect(filepath.Join(layers.Path, "test-name")).To(BeADirectory())

					files, err := filepath.Glob(filepath.Join(layers.Path, "test-name", "*"))
					Expect(err).NotTo(HaveOccurred())

					Expect(files).To(BeEmpty())
				})
			})
		})

		context("could not remove files in layer", func() {
			it.Before(func() {
				Expect(os.Chmod(layers.Path, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(layers.Path, 0777)).To(Succeed())
			})

			it("return an error", func() {
				layer := libcnb.Layer{
					Name: "some-layer",
					Path: filepath.Join(layers.Path, "some-layer"),
				}

				_, err := layer.Reset()
				Expect(err).To(MatchError(ContainSubstring("error could not remove file: ")))
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})

	context("Layers", func() {
		it.Before(func() {
			var err error
			path, err = ioutil.TempDir("", "layers")
			Expect(err).NotTo(HaveOccurred())

			layers = libcnb.Layers{Path: path}
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("initializes layer", func() {
			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.LayerTypes.Build).To(BeFalse())
			Expect(l.LayerTypes.Cache).To(BeFalse())
			Expect(l.LayerTypes.Launch).To(BeFalse())
			Expect(l.Metadata).To(BeNil())
			Expect(l.Name).To(Equal("test-name"))
			Expect(l.Path).To(Equal(filepath.Join(path, "test-name")))
			Expect(l.BuildEnvironment).To(Equal(libcnb.Environment{}))
			Expect(l.LaunchEnvironment).To(Equal(libcnb.Environment{}))
			Expect(l.SharedEnvironment).To(Equal(libcnb.Environment{}))
			Expect(l.Profile).To(Equal(libcnb.Profile{}))
		})

		it("generates SBOM paths", func() {
			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.Path).To(Equal(filepath.Join(path, "test-name")))
			Expect(layers.BuildSBOMPath(libcnb.CycloneDXJSON)).To(Equal(filepath.Join(path, "build.sbom.cdx.json")))
			Expect(layers.BuildSBOMPath(libcnb.SPDXJSON)).To(Equal(filepath.Join(path, "build.sbom.spdx.json")))
			Expect(layers.BuildSBOMPath(libcnb.SyftJSON)).To(Equal(filepath.Join(path, "build.sbom.syft.json")))
			Expect(layers.LaunchSBOMPath(libcnb.SyftJSON)).To(Equal(filepath.Join(path, "launch.sbom.syft.json")))
			Expect(l.SBOMPath(libcnb.SyftJSON)).To(Equal(filepath.Join(path, "test-name.sbom.syft.json")))
		})

		it("maps from string to SBOM Format", func() {
			fmt, err := libcnb.SBOMFormatFromString("cdx.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(fmt).To(Equal(libcnb.CycloneDXJSON))

			fmt, err = libcnb.SBOMFormatFromString("spdx.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(fmt).To(Equal(libcnb.SPDXJSON))

			fmt, err = libcnb.SBOMFormatFromString("syft.json")
			Expect(err).ToNot(HaveOccurred())
			Expect(fmt).To(Equal(libcnb.SyftJSON))

			fmt, err = libcnb.SBOMFormatFromString("foobar.json")
			Expect(err).To(MatchError("unable to translate from foobar.json to SBOMFormat"))
			Expect(fmt).To(Equal(libcnb.UnknownFormat))
		})

		it("reads existing 0.5 metadata", func() {
			Expect(ioutil.WriteFile(
				filepath.Join(path, "test-name.toml"),
				[]byte(`
launch = true
build = false

[metadata]
test-key = "test-value"
		`),
				0600),
			).To(Succeed())

			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.Metadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
			Expect(l.Launch).To(BeTrue())
			Expect(l.Build).To(BeFalse())
		})

		it("reads existing 0.6 metadata", func() {
			Expect(ioutil.WriteFile(
				filepath.Join(path, "test-name.toml"),
				[]byte(`
[types]
launch = true
build = false

[metadata]
test-key = "test-value"
		`),
				0600),
			).To(Succeed())

			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.Metadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
			Expect(l.Launch).To(BeTrue())
			Expect(l.Build).To(BeFalse())
		})

		it("reads existing 0.6 metadata with launch, build and cache all false", func() {
			Expect(ioutil.WriteFile(
				filepath.Join(path, "test-name.toml"),
				[]byte(`
[types]
launch = false
build = false
cache = false

[metadata]
test-key = "test-value"
		`),
				0600),
			).To(Succeed())

			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.Metadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
			Expect(l.Launch).To(BeFalse())
			Expect(l.Build).To(BeFalse())
			Expect(l.Cache).To(BeFalse())
		})
	})
}
