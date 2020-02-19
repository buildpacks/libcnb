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

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
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

	context("Binding", func() {
		it("creates an empty binding", func() {
			Expect(libcnb.NewBinding()).To(Equal(libcnb.Binding{
				Metadata: map[string]string{},
				Secret:   map[string]string{},
			}))
		})

		it("creates a binding from a path", func() {
			Expect(os.MkdirAll(filepath.Join(path, "metadata"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "metadata", "test-metadata-key"), []byte("test-metadata-value"), 0644)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "secret"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "secret", "test-secret-key"), []byte("test-secret-value"), 0644)).To(Succeed())

			Expect(libcnb.NewBindingFromPath(path)).To(Equal(libcnb.Binding{
				Metadata: map[string]string{
					"test-metadata-key": "test-metadata-value",
				},
				Secret: map[string]string{
					"test-secret-key": "test-secret-value",
				},
			}))
		})

		it("sanitizes secrets", func() {
			Expect(os.MkdirAll(filepath.Join(path, "metadata"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "metadata", "test-metadata-key"), []byte("test-metadata-value"), 0644)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "secret"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(path, "secret", "test-secret-key"), []byte("test-secret-value"), 0644)).To(Succeed())

			b, err := libcnb.NewBindingFromPath(path)
			Expect(err).NotTo(HaveOccurred())

			Expect(b.String()).To(Equal("{Metadata: map[test-metadata-key:test-metadata-value] Secret: [test-secret-key]}"))

		})
	})

	context("Bindings", func() {
		it("creates a bindings from a path", func() {
			Expect(os.MkdirAll(filepath.Join(path, "alpha"), 0755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(path, "bravo"), 0755)).To(Succeed())

			Expect(libcnb.NewBindingsFromPath(path)).To(Equal(libcnb.Bindings{
				"alpha": libcnb.NewBinding(),
				"bravo": libcnb.NewBinding(),
			}))
		})

		context("from environment", func() {
			it.Before(func() {
				Expect(os.Setenv("BINDINGS_TEST", `
[alpha]
[alpha.metadata]
kind = "alpha-kind"
provider = "alpha-provider"
tags = "alpha-tag-1\nalpha-tag-2"
other-key = "alpha-other-value"

[alpha.secret]
secret-key = "secret-value"

[bravo]
[bravo.metadata]
kind = "bravo-kind"
provider = "bravo-provider"
tags = "bravo-tag-1\nbravo-tag-2"
other-key = "bravo-other-value"
`))
			})

			it.After(func() {
				Expect(os.Unsetenv("BINDINGS_TEST"))
			})

			it("returns empty bindings if environment variable is not set", func() {
				Expect(libcnb.NewBindingsFromEnvironment("BINDINGS_UNSET")).To(Equal(libcnb.Bindings{}))
			})

			it("creates bindings from an environment variable", func() {
				Expect(libcnb.NewBindingsFromEnvironment("BINDINGS_TEST")).To(Equal(libcnb.Bindings{
					"alpha": libcnb.Binding{
						Metadata: map[string]string{
							libcnb.BindingKind:     "alpha-kind",
							libcnb.BindingProvider: "alpha-provider",
							libcnb.BindingTags:     "alpha-tag-1\nalpha-tag-2",
							"other-key":            "alpha-other-value",
						},
						Secret: map[string]string{
							"secret-key": "secret-value",
						},
					},
					"bravo": libcnb.Binding{
						Metadata: map[string]string{
							libcnb.BindingKind:     "bravo-kind",
							libcnb.BindingProvider: "bravo-provider",
							libcnb.BindingTags:     "bravo-tag-1\nbravo-tag-2",
							"other-key":            "bravo-other-value",
						},
						Secret: map[string]string{},
					},
				}))
			})
		})
	})
}
