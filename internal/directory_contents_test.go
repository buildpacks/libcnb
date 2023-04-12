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

func testDirectoryContents(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
	)

	it.Before(func() {
		var err error
		path, err = ioutil.TempDir("", "directory-contents")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	it("lists empty directory contents", func() {
		Expect(internal.DirectoryContents{path}.Get()).To(Equal([]string{"."}))
	})

	it("lists directory contents", func() {
		f, err := os.Create(filepath.Join(path, "test-file"))
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()

		Expect(internal.DirectoryContents{path}.Get()).To(Equal([]string{".", "test-file"}))
	})
}
