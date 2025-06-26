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
		platformPath = t.TempDir()
		path = filepath.Join(platformPath, "bindings")
	})

	context("Cloudfoundry VCAP_SERVICES", func() {
		context("Build", func() {
			it("creates a bindings from VCAP_SERVICES", func() {
				content, err := os.ReadFile("testdata/vcap_services.json")
				Expect(err).NotTo(HaveOccurred())
				t.Setenv(libcnb.EnvVcapServices, string(content))

				bindings, err := libcnb.NewBindingsForBuild("")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(ConsistOf(libcnb.Bindings{
					{
						Name:     "elephantsql-binding-c6c60",
						Type:     "elephantsql-type",
						Provider: "elephantsql-provider",
						Secret: map[string]string{
							"bool": "true",
							"int":  "1",
							"uri":  "postgres://exampleuser:examplepass@postgres.example.com:5432/exampleuser",
						},
					},
					{
						Name:     "mysendgrid",
						Type:     "sendgrid-type",
						Provider: "sendgrid-provider",
						Secret: map[string]string{
							"username": "QvsXMbJ3rK",
							"password": "HCHMOYluTv",
							"hostname": "smtp.example.com",
						},
					},
					{
						Name:     "postgres",
						Type:     "postgres",
						Provider: "postgres",
						Secret: map[string]string{
							"urls":     "{\"example\":\"http://example.com\"}",
							"username": "foo",
							"password": "bar",
						},
					},
					{
						Name:     "my-custom-binding",
						Type:     "custom-type",
						Provider: "user-provided",
						Secret: map[string]string{
							"username": "foo",
							"password": "bar",
							"type":     "custom-type",
						},
					},
				}))
			})

			it("creates empty bindings from empty VCAP_SERVICES", func() {
				t.Setenv(libcnb.EnvVcapServices, "{}")

				bindings, err := libcnb.NewBindingsForBuild("")
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(HaveLen(0))
			})
		})

		context("Launch", func() {
			it("creates a bindings from VCAP_SERVICES", func() {
				content, err := os.ReadFile("testdata/vcap_services.json")
				Expect(err).NotTo(HaveOccurred())
				t.Setenv(libcnb.EnvVcapServices, string(content))

				bindings, err := libcnb.NewBindingsForLaunch()
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(HaveLen(4))
				types := []string{bindings[0].Type, bindings[1].Type, bindings[2].Type}
				Expect(types).To(ContainElements("elephantsql-type", "sendgrid-type", "postgres"))
			})

			it("creates empty bindings from empty VCAP_SERVICES", func() {
				t.Setenv(libcnb.EnvVcapServices, "{}")

				bindings, err := libcnb.NewBindingsForLaunch()
				Expect(err).NotTo(HaveOccurred())

				Expect(bindings).To(HaveLen(0))
			})
		})
	})

	context("CNB Bindings", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(path, "alpha", "metadata"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "metadata", "kind"), []byte("test-kind"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "metadata", "provider"), []byte("test-provider"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "metadata", "test-metadata-key"), []byte("test-metadata-value"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "metadata", "test-metadata-key-trimmed"), []byte(" test-metadata-value-trimmed \n"), 0600)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "alpha", "secret"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "secret", "test-secret-key"), []byte("test-secret-value"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "alpha", "secret", "test-secret-key-trimmed"), []byte(" test-secret-value-trimmed \n"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(path, "bravo", "metadata"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "metadata", "kind"), []byte("test-kind"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "metadata", "provider"), []byte("test-provider"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "metadata", "test-metadata-key"), []byte("test-metadata-value"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "metadata", "test-metadata-key-trimmed"), []byte(" test-metadata-value-trimmed \n"), 0600)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "bravo", "secret"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "secret", "test-secret-key"), []byte("test-secret-value"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, "bravo", "secret", "test-secret-key-trimmed"), []byte(" test-secret-value-trimmed \n"), 0600)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(path, ".hidden", "metadata"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, ".hidden", "metadata", "kind"), []byte("test-kind"), 0600)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(path, ".hiddenFile"), []byte("test-kind"), 0600)).To(Succeed())
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
					Name:     filepath.Base(path),
					Path:     path,
					Type:     "test-kind",
					Provider: "test-provider",
					Secret: map[string]string{
						"test-metadata-key":         "test-metadata-value",
						"test-metadata-key-trimmed": "test-metadata-value-trimmed",
						"test-secret-key":           "test-secret-value",
						"test-secret-key-trimmed":   "test-secret-value-trimmed",
					},
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

				Expect(b.String()).To(Equal(fmt.Sprintf("{Name: alpha Path: %s Type: test-kind Provider: test-provider Secret: [test-metadata-key test-metadata-key-trimmed test-secret-key test-secret-key-trimmed]}", path)))
			})
		})

		context("Bindings", func() {
			it("creates a bindings from a path", func() {
				Expect(libcnb.NewBindingsFromPath(path)).To(Equal(libcnb.Bindings{
					libcnb.Binding{
						Name:     "alpha",
						Path:     filepath.Join(path, "alpha"),
						Type:     "test-kind",
						Provider: "test-provider",
						Secret: map[string]string{
							"test-metadata-key":         "test-metadata-value",
							"test-metadata-key-trimmed": "test-metadata-value-trimmed",
							"test-secret-key":           "test-secret-value",
							"test-secret-key-trimmed":   "test-secret-value-trimmed",
						},
					},
					libcnb.Binding{
						Name:     "bravo",
						Path:     filepath.Join(path, "bravo"),
						Type:     "test-kind",
						Provider: "test-provider",
						Secret: map[string]string{
							"test-metadata-key":         "test-metadata-value",
							"test-metadata-key-trimmed": "test-metadata-value-trimmed",
							"test-secret-key":           "test-secret-value",
							"test-secret-key-trimmed":   "test-secret-value-trimmed",
						},
					},
				}))
			})

			it("returns empty bindings if environment variable is not set", func() {
				Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{}))
			})

			context("from environment", func() {
				it.Before(func() {
					Expect(os.Setenv(libcnb.EnvCNBBindings, path))
				})

				it.After(func() {
					Expect(os.Unsetenv(libcnb.EnvCNBBindings))
				})

				it("creates bindings from path in $CNB_BINDINGS", func() {
					Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{
						libcnb.Binding{
							Name:     "alpha",
							Path:     filepath.Join(path, "alpha"),
							Type:     "test-kind",
							Provider: "test-provider",
							Secret: map[string]string{
								"test-metadata-key":         "test-metadata-value",
								"test-metadata-key-trimmed": "test-metadata-value-trimmed",
								"test-secret-key":           "test-secret-value",
								"test-secret-key-trimmed":   "test-secret-value-trimmed",
							},
						},
						libcnb.Binding{
							Name:     "bravo",
							Path:     filepath.Join(path, "bravo"),
							Type:     "test-kind",
							Provider: "test-provider",
							Secret: map[string]string{
								"test-metadata-key":         "test-metadata-value",
								"test-metadata-key-trimmed": "test-metadata-value-trimmed",
								"test-secret-key":           "test-secret-value",
								"test-secret-key-trimmed":   "test-secret-value-trimmed",
							},
						},
					}))
				})
			})
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

			it("returns empty bindings if environment variable is not set", func() {
				Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{}))
			})

			context("from environment", func() {
				it.Before(func() {
					Expect(os.Setenv(libcnb.EnvServiceBindings, path))
				})

				it.After(func() {
					Expect(os.Unsetenv(libcnb.EnvServiceBindings))
				})

				it("creates bindings from path in SERVICE_BINDING_ROOT", func() {
					Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{
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

			context("from environment or path", func() {
				context("when SERVICE_BINDING_ROOT is defined but CNB_BINDINGS or the passed path does not exist", func() {
					it.Before(func() {
						Expect(os.Setenv(libcnb.EnvServiceBindings, path))
						Expect(os.Setenv(libcnb.EnvCNBBindings, "does not exist"))
					})

					it.After(func() {
						Expect(os.Unsetenv(libcnb.EnvServiceBindings))
						Expect(os.Unsetenv(libcnb.EnvCNBBindings))
					})

					it("creates bindings from path in SERVICE_BINDING_ROOT", func() {
						Expect(libcnb.NewBindingsForBuild("random-path-that-does-not-exist")).To(Equal(libcnb.Bindings{
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

				context("when CNB_BINDINGS is defined but the path does not exist", func() {
					it.Before(func() {
						Expect(os.Setenv(libcnb.EnvCNBBindings, path))
					})

					it.After(func() {
						Expect(os.Unsetenv(libcnb.EnvCNBBindings))
					})

					it("creates bindings from path in CNB_BINDINGS", func() {
						Expect(libcnb.NewBindingsForBuild("random-path-that-does-not-exist")).To(Equal(libcnb.Bindings{
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

				context("when SERVICE_BINDING_ROOT and CNB_BINDINGS is not defined but the path exists", func() {
					it("creates bindings from the given path", func() {
						Expect(libcnb.NewBindingsForBuild(platformPath)).To(Equal(libcnb.Bindings{
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

				context("when no valid binding variable is set", func() {
					it("returns an an empty binding", func() {
						Expect(libcnb.NewBindingsForBuild("does-not-exist")).To(Equal(libcnb.Bindings{}))
					})
				})
			})
		})
	})
}
