/*
 * Copyright 2018-2022 the original author or authors.
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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb/internal"
)

func testDirectoryContentsWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
		buf  bytes.Buffer
	)

	it.Before(func() {
		var err error
		path, err = os.MkdirTemp("", "directory-contents")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(path)).To(Succeed())
	})

	context("directory content formats", func() {
		fm := internal.NewPlainDirectoryContentFormatter()

		it("formats title", func() {
			Expect(fm.Title("foo")).To(Equal("foo:\n"))
		})

		it("formats a file", func() {
			cwd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			info, err := os.Stat(cwd)
			Expect(err).ToNot(HaveOccurred())

			fm.RootPath(filepath.Dir(cwd))

			Expect(fm.File(cwd, info)).To(Equal(fmt.Sprintf("%s\n", filepath.Base(cwd))))
		})
	})

	it("lists empty directory contents", func() {
		fm := internal.NewPlainDirectoryContentFormatter()
		dc := internal.NewDirectoryContentsWriter(fm, &buf)

		Expect(dc.Write("title", path)).To(Succeed())
		Expect(buf.String()).To(Equal("title:\n.\n"))
	})

	it("lists directory contents", func() {
		f, err := os.Create(filepath.Join(path, "test-file"))
		Expect(err).NotTo(HaveOccurred())
		defer f.Close()

		fm := internal.NewPlainDirectoryContentFormatter()
		dc := internal.NewDirectoryContentsWriter(fm, &buf)

		Expect(dc.Write("title", path)).To(Succeed())
		Expect(buf.String()).To(Equal("title:\n.\ntest-file\n"))
	})
}
