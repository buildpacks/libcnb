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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb/v2/internal"
)

func testFormatters(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		path string
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
}
