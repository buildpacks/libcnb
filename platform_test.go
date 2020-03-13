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

		it("returns empty bindings if environment variable is not set", func() {
			Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{}))
		})

		context("from environment", func() {
			it.Before(func() {
				Expect(os.MkdirAll(filepath.Join(path, "alpha"), 0755)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(path, "bravo"), 0755)).To(Succeed())

				Expect(os.Setenv("CNB_BINDINGS", path))
			})

			it.After(func() {
				Expect(os.Unsetenv("CNB_BINDINGS"))
			})

			it("creates bindings from path in $CNB_BINDINGS", func() {
				Expect(libcnb.NewBindingsFromEnvironment()).To(Equal(libcnb.Bindings{
					"alpha": libcnb.NewBinding(),
					"bravo": libcnb.NewBinding(),
				}))
			})
		})
	})
}
