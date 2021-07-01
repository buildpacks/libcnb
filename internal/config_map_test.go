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

package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb/internal"
)

func testConfigMap(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "config-map")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("returns an empty ConfigMap when directory does not exist", func() {
		Expect(os.RemoveAll(path)).To(Succeed())

		cm, err := internal.NewConfigMapFromPath(path)
		Expect(err).NotTo(HaveOccurred())

		Expect(cm).To(Equal(internal.ConfigMap{}))
	})

	it("loads the ConfigMap from a directory", func() {
		Expect(ioutil.WriteFile(filepath.Join(path, "test-key"), []byte("test-value"), 0600)).To(Succeed())

		cm, err := internal.NewConfigMapFromPath(path)
		Expect(err).NotTo(HaveOccurred())

		Expect(cm).To(Equal(internal.ConfigMap{"test-key": "test-value"}))
	})

	it("ignores dirs and follows symlinks", func() {
		// this is necessary to support bindings mounted as k8s config maps & secrets
		Expect(os.MkdirAll(filepath.Join(path, ".hidden"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(
			filepath.Join(path, ".hidden", "test-key"),
			[]byte("test-value"),
			0600,
		)).To(Succeed())
		Expect(os.Symlink(
			filepath.Join(".hidden", "test-key"),
			filepath.Join(path, "test-key"),
		)).To(Succeed())

		cm, err := internal.NewConfigMapFromPath(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(cm).To(Equal(internal.ConfigMap{"test-key": "test-value"}))
	})

	it("ignores hidden files", func() {
		Expect(ioutil.WriteFile(filepath.Join(path, ".hidden-key"), []byte("hidden-value"), 0600)).To(Succeed())

		cm, err := internal.NewConfigMapFromPath(path)
		Expect(err).NotTo(HaveOccurred())

		Expect(cm).To(BeEmpty())
	})
}
