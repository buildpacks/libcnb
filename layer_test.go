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

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb"
)

func testLayer(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layers libcnb.Layers
		path   string
	)

	context("Exec", func() {
		var exec libcnb.Exec

		it.Before(func() {
			exec = libcnb.Exec{Path: "test-path"}
		})

		it("returns filename", func() {
			Expect(exec.FilePath("test-name")).To(Equal(filepath.Join("test-path", "test-name")))
		})

		it("returns process-specific filename", func() {
			Expect(exec.ProcessFilePath("test-process", "test-name")).
				To(Equal(filepath.Join("test-path", "test-process", "test-name")))
		})
	})

	context("Profile", func() {
		var profile libcnb.Profile

		it.Before(func() {
			profile = libcnb.Profile{}
		})

		it("adds content", func() {
			profile.Add("test-name", "test-value")
			Expect(profile).To(Equal(libcnb.Profile{"test-name": "test-value"}))
		})

		it("adds formatted content", func() {
			profile.Addf("test-name", "test-%s", "value")
			Expect(profile).To(Equal(libcnb.Profile{"test-name": "test-value"}))
		})

		it("adds process-specific content", func() {
			profile.ProcessAdd("test-process", "test-name", "test-value")
			Expect(profile).To(Equal(libcnb.Profile{filepath.Join("test-process", "test-name"): "test-value"}))
		})

		it("adds process-specific formatted content", func() {
			profile.ProcessAddf("test-process", "test-name", "test-%s", "value")
			Expect(profile).To(Equal(libcnb.Profile{filepath.Join("test-process", "test-name"): "test-value"}))
		})
	})

	context("Layers", func() {

		it.Before(func() {
			var err error
			path, err = ioutil.TempDir("", "layers")
			Expect(err).NotTo(HaveOccurred())

			layers = libcnb.Layers{Path: path}
		})

		it.After(func() {
			Expect(os.RemoveAll(path)).To(Succeed())
		})

		it("initializes layer", func() {
			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.LayerTypes.Build).To(BeFalse())
			Expect(l.LayerTypes.Cache).To(BeFalse())
			Expect(l.LayerTypes.Launch).To(BeFalse())
			Expect(l.Metadata).To(BeNil())
			Expect(l.Name).To(Equal("test-name"))
			Expect(l.Path).To(Equal(filepath.Join(path, "test-name")))
			Expect(l.BuildEnvironment).To(Equal(libcnb.Environment{}))
			Expect(l.LaunchEnvironment).To(Equal(libcnb.Environment{}))
			Expect(l.SharedEnvironment).To(Equal(libcnb.Environment{}))
			Expect(l.Profile).To(Equal(libcnb.Profile{}))
		})

		it("reads existing metadata", func() {
			Expect(ioutil.WriteFile(
				filepath.Join(path, "test-name.toml"),
				[]byte(`
[types]
launch = true
build = false

[metadata]
test-key = "test-value"
		`),
				0644),
			).To(Succeed())

			l, err := layers.Layer("test-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(l.Metadata).To(Equal(map[string]interface{}{"test-key": "test-value"}))
			Expect(l.Launch).To(BeTrue())
			Expect(l.Build).To(BeFalse())
		})
	})
}
