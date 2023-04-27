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
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb"
)

func testPlatform(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path, platformPath string
	)

	it.Before(func() {
		var err error
		platformPath, err = os.MkdirTemp("", "platform")
		path = filepath.Join(platformPath, "bindings")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("Cloudfoundry VCAP_SERVICES", func() {
		it("creates a bindings from VCAP_SERVICES", func() {
			content, err := os.ReadFile("testdata/vcap_services.json")
			Expect(err).NotTo(HaveOccurred())
			t.Setenv(libcnb.EnvVcapServices, string(content))

			bindings, err := libcnb.NewBindings("")
			Expect(err).NotTo(HaveOccurred())

			Expect(bindings).To(HaveLen(2))
			Expect(bindings[0].Type).To(Equal("elephantsql"))
			Expect(bindings[1].Type).To(Equal("sendgrid"))
		})

		it("creates empty bindings from empty VCAP_SERVICES", func() {
			t.Setenv(libcnb.EnvVcapServices, "{}")

			bindings, err := libcnb.NewBindings("")
			Expect(err).NotTo(HaveOccurred())

			Expect(bindings).To(HaveLen(0))
		})
	})

	context("Kubernetes Service Bindings", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(path, "alpha"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "type"), []byte("test-type"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "provider"), []byte("test-provider"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "test-secret-key"), []byte("test-secret-value"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(path, "bravo"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "type"), []byte("test-type"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "provider"), []byte("test-provider"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "test-secret-key"), []byte("test-secret-value"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(path, ".hidden", "metadata"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, ".hidden", "metadata", "kind"), []byte("test-kind"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, ".hiddenFile"), []byte("test-kind"), 0600)).To(Succeed())
		})

		context("Binding", func() {
			it("creates an empty binding", func() {
				Expect(libcnb.NewBinding("test-name", "test-path", map[string]string{
					libcnb.BindingType:     "test-type",
					libcnb.BindingProvider: "test-provider",
					"test-key":             "test-value",
				})).To(Equal(libcnb.Binding{
					Name:     "test-name",
					Path:     "test-path",
					Type:     "test-type",
					Provider: "test-provider",
					Secret: map[string]string{
						"test-key": "test-value",
					},
				}))
			})

			it("creates a binding from a path", func() {
				path := filepath.Join(path, "alpha")

				binding, err := libcnb.NewBindingFromPath(path)
				Expect(binding, err).To(Equal(libcnb.Binding{
					Name:     filepath.Base(path),
					Path:     path,
					Type:     "test-type",
					Provider: "test-provider",
					Secret:   map[string]string{"test-secret-key": "test-secret-value"},
				}))

				secretFilePath, ok := binding.SecretFilePath("test-secret-key")
				Expect(ok).To(BeTrue())
				Expect(secretFilePath).To(Equal(filepath.Join(path, "test-secret-key")))
			})

			it("sanitizes secrets", func() {
				path := filepath.Join(path, "alpha")

				b, err := libcnb.NewBindingFromPath(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(b.String()).To(Equal(fmt.Sprintf("{Name: alpha Path: %s Type: test-type Provider: test-provider Secret: [test-secret-key]}", path)))
			})
		})

		context("Bindings", func() {
			it("creates a bindings from a path", func() {
				Expect(libcnb.NewBindingsFromPath(path)).To(Equal(libcnb.Bindings{
					libcnb.Binding{
						Name:     "alpha",
						Path:     filepath.Join(path, "alpha"),
						Type:     "test-type",
						Provider: "test-provider",
						Secret:   map[string]string{"test-secret-key": "test-secret-value"},
					},
					libcnb.Binding{
						Name:     "bravo",
						Path:     filepath.Join(path, "bravo"),
						Type:     "test-type",
						Provider: "test-provider",
						Secret:   map[string]string{"test-secret-key": "test-secret-value"},
					},
				}))
			})

			it("creates an empty binding if path does not exist", func() {
				Expect(libcnb.NewBindingsFromPath("/path/doesnt/exist")).To(Equal(libcnb.Bindings{}))
			})

			it("returns empty bindings if SERVICE_BINDING_ROOT and CNB_PLATFORM_DIR are not set and /platform/bindings does not exist", func() {
				Expect(libcnb.NewBindings(libcnb.DefaultPlatformBindingsLocation)).To(Equal(libcnb.Bindings{}))
			})

			context("from environment", func() {
				it.After(func() {
					Expect(os.Unsetenv(libcnb.EnvServiceBindings))
					Expect(os.Unsetenv("CNB_PLATFORM_DIR"))
				})

				it("creates bindings from path in SERVICE_BINDING_ROOT if both set", func() {
					Expect(os.Setenv(libcnb.EnvServiceBindings, path))
					Expect(os.Setenv("CNB_PLATFORM_DIR", "/does/not/exist"))

					Expect(libcnb.NewBindings(libcnb.DefaultPlatformBindingsLocation)).To(Equal(libcnb.Bindings{
						libcnb.Binding{
							Name:     "alpha",
							Path:     filepath.Join(path, "alpha"),
							Type:     "test-type",
							Provider: "test-provider",
							Secret:   map[string]string{"test-secret-key": "test-secret-value"},
						},
						libcnb.Binding{
							Name:     "bravo",
							Path:     filepath.Join(path, "bravo"),
							Type:     "test-type",
							Provider: "test-provider",
							Secret:   map[string]string{"test-secret-key": "test-secret-value"},
						},
					}))
				})

				it("creates bindings from path in SERVICE_BINDING_ROOT if SERVICE_BINDING_ROOT not set", func() {
					Expect(os.Setenv("CNB_PLATFORM_DIR", filepath.Dir(path)))

					Expect(libcnb.NewBindings(libcnb.DefaultPlatformBindingsLocation)).To(Equal(libcnb.Bindings{
						libcnb.Binding{
							Name:     "alpha",
							Path:     filepath.Join(path, "alpha"),
							Type:     "test-type",
							Provider: "test-provider",
							Secret:   map[string]string{"test-secret-key": "test-secret-value"},
						},
						libcnb.Binding{
							Name:     "bravo",
							Path:     filepath.Join(path, "bravo"),
							Type:     "test-type",
							Provider: "test-provider",
							Secret:   map[string]string{"test-secret-key": "test-secret-value"},
						},
					}))
				})
			})
		})
	})
}
