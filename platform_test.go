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

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb"
)

func testPlatform(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "bindings")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("CNB Bindings", func() {

		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(path, "alpha", "metadata"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "alpha", "metadata", "test-metadata-key"), []byte("test-metadata-value"), 0644)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "alpha", "secret"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "alpha", "secret", "test-secret-key"), []byte("test-secret-value"), 0644)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(path, "bravo", "metadata"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "bravo", "metadata", "test-metadata-key"), []byte("test-metadata-value"), 0644)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "bravo", "secret"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "bravo", "secret", "test-secret-key"), []byte("test-secret-value"), 0644)).To(Succeed())
		})

		context("Binding", func() {
			it("creates an empty binding", func() {
				Expect(libcnb.NewBinding("test-name", "test-path", map[string]string{
					libcnb.BindingKind:     "test-kind",
					libcnb.BindingProvider: "test-provider",
					"test-key":             "test-value",
				})).To(Equal(libcnb.Binding{
					Name:     "test-name",
					Path:     "test-path",
					Type:     "test-kind",
					Provider: "test-provider",
					Secret:   map[string]string{"test-key": "test-value"},
				}))
			})

			it("creates a binding from a path", func() {
				path := filepath.Join(path, "alpha")

				binding, err := libcnb.NewBindingFromPath(path)
				Expect(binding, err).To(Equal(libcnb.Binding{
					Name: filepath.Base(path),
					Secret: map[string]string{
						"test-metadata-key": "test-metadata-value",
						"test-secret-key":   "test-secret-value",
					},
					Path: path,
				}))

				metadataFilePath, ok := binding.SecretFilePath("test-metadata-key")
				Expect(ok).To(BeTrue())
				Expect(metadataFilePath).To(Equal(filepath.Join(path, "metadata", "test-metadata-key")))

				secretFilePath, ok := binding.SecretFilePath("test-secret-key")
				Expect(ok).To(BeTrue())
				Expect(secretFilePath).To(Equal(filepath.Join(path, "secret", "test-secret-key")))
			})

			it("sanitizes secrets", func() {
				path := filepath.Join(path, "alpha")

				b, err := libcnb.NewBindingFromPath(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(b.String()).To(Equal(fmt.Sprintf("{Path: %s Secret: [test-metadata-key test-secret-key]}", path)))
			})
		})

		context("Bindings", func() {
			it("creates a bindings from a path", func() {
				Expect(libcnb.NewBindingsFromPath(path)).To(Equal(libcnb.Bindings{
					libcnb.Binding{
						Name: "alpha",
						Secret: map[string]string{
							"test-metadata-key": "test-metadata-value",
							"test-secret-key":   "test-secret-value",
						},
						Path: filepath.Join(path, "alpha"),
					},
					libcnb.Binding{
						Name: "bravo",
						Secret: map[string]string{
							"test-metadata-key": "test-metadata-value",
							"test-secret-key":   "test-secret-value",
						},
						Path: filepath.Join(path, "bravo"),
					},
				}))
			})

			it("returns empty bindings if environment variable is not set", func() {
				Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{}))
			})

			context("from environment", func() {
				it.Before(func() {
					Expect(os.Setenv("CNB_BINDINGS", path))
				})

				it.After(func() {
					Expect(os.Unsetenv("CNB_BINDINGS"))
				})

				it("creates bindings from path in $CNB_BINDINGS", func() {
					Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{
						libcnb.Binding{
							Name: "alpha",
							Secret: map[string]string{
								"test-metadata-key": "test-metadata-value",
								"test-secret-key":   "test-secret-value",
							},
							Path: filepath.Join(path, "alpha"),
						},
						libcnb.Binding{
							Name: "bravo",
							Secret: map[string]string{
								"test-metadata-key": "test-metadata-value",
								"test-secret-key":   "test-secret-value",
							},
							Path: filepath.Join(path, "bravo"),
						},
					}))
				})
			})
		})
	})

	context("Kubernetes Service Bindings", func() {

		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(path, "alpha"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "alpha", "test-secret-key"), []byte("test-secret-value"), 0644)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(path, "bravo"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "bravo", "test-secret-key"), []byte("test-secret-value"), 0644)).To(Succeed())
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
					Name:   filepath.Base(path),
					Secret: map[string]string{"test-secret-key": "test-secret-value"},
					Path:   path,
				}))

				secretFilePath, ok := binding.SecretFilePath("test-secret-key")
				Expect(ok).To(BeTrue())
				Expect(secretFilePath).To(Equal(filepath.Join(path, "test-secret-key")))
			})

			it("sanitizes secrets", func() {
				path := filepath.Join(path, "alpha")

				b, err := libcnb.NewBindingFromPath(path)
				Expect(err).NotTo(HaveOccurred())

				Expect(b.String()).To(Equal(fmt.Sprintf("{Path: %s Secret: [test-secret-key]}", path)))
			})
		})

		context("Bindings", func() {
			it("creates a bindings from a path", func() {
				Expect(libcnb.NewBindingsFromPath(path)).To(Equal(libcnb.Bindings{
					libcnb.Binding{
						Name:   "alpha",
						Secret: map[string]string{"test-secret-key": "test-secret-value"},
						Path:   filepath.Join(path, "alpha"),
					},
					libcnb.Binding{
						Name:   "bravo",
						Secret: map[string]string{"test-secret-key": "test-secret-value"},
						Path:   filepath.Join(path, "bravo"),
					},
				}))
			})

			it("returns empty bindings if environment variable is not set", func() {
				Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{}))
			})

			context("from environment", func() {
				it.Before(func() {
					Expect(os.Setenv("SERVICE_BINDING_ROOT", path))
				})

				it.After(func() {
					Expect(os.Unsetenv("SERVICE_BINDING_ROOT"))
				})

				it("creates bindings from path in SERVICE_BINDING_ROOT", func() {
					Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{
						libcnb.Binding{
							Name:   "alpha",
							Secret: map[string]string{"test-secret-key": "test-secret-value"},
							Path:   filepath.Join(path, "alpha"),
						},
						libcnb.Binding{
							Name:   "bravo",
							Secret: map[string]string{"test-secret-key": "test-secret-value"},
							Path:   filepath.Join(path, "bravo"),
						},
					}))
				})
			})
		})
	})

}
