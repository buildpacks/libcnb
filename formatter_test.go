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

func testFormatter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("ApplicationPathFormatter", func() {
		var (
			app string
		)

		it.Before(func() {
			var err error
			app, err = ioutil.TempDir("", "application-path-formatter")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(app)).To(Succeed())
		})

		it("lists empty directory contents", func() {
			Expect(libcnb.ApplicationPathFormatter(app).String()).To(Equal("Application contents: [.]"))
		})

		it("lists directory contents", func() {
			f, err := os.Create(filepath.Join(app, "test-file"))
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			Expect(libcnb.ApplicationPathFormatter(app).String()).To(Equal("Application contents: [. test-file]"))
		})
	})

	context("BuildpackPathFormatter", func() {
		var (
			bp string
		)

		it.Before(func() {
			var err error
			bp, err = ioutil.TempDir("", "buildpack-path-formatter")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(bp)).To(Succeed())
		})

		it("lists empty directory contents", func() {
			Expect(libcnb.BuildpackPathFormatter(bp).String()).To(Equal("Buildpack contents: [.]"))
		})

		it("lists directory contents", func() {
			f, err := os.Create(filepath.Join(bp, "test-file"))
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			Expect(libcnb.BuildpackPathFormatter(bp).String()).To(Equal("Buildpack contents: [. test-file]"))
		})
	})

	context("PlatformFormatter", func() {
		var (
			plat libcnb.Platform
		)

		it.Before(func() {
			var err error
			plat.Path, err = ioutil.TempDir("", "platform-formatter")
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(os.RemoveAll(plat.Path)).To(Succeed())
		})

		it("lists empty directory contents", func() {
			Expect(libcnb.PlatformFormatter(plat).String()).To(Equal("Platform contents: [.]"))
		})

		it("lists directory contents", func() {
			f, err := os.Create(filepath.Join(plat.Path, "test-file"))
			Expect(err).NotTo(HaveOccurred())
			defer f.Close()

			Expect(libcnb.PlatformFormatter(plat).String()).To(Equal("Platform contents: [. test-file]"))
		})

	})

}
