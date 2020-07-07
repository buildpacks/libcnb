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

func testTOMLWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		parent     string
		path       string
		tomlWriter internal.TOMLWriter
	)

	it.Before(func() {
		var err error
		parent, err = ioutil.TempDir("", "toml-writer")
		Expect(err).NotTo(HaveOccurred())

		path = filepath.Join(parent, "text.toml")
	})

	it.After(func() {
		Expect(os.RemoveAll(parent)).To(Succeed())
	})

	it("writes the contents of a given object out to a .toml file", func() {
		err := tomlWriter.Write(path, map[string]string{
			"some-field":  "some-value",
			"other-field": "other-value",
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.ReadFile(path)).To(internal.MatchTOML(`
some-field = "some-value"
other-field = "other-value"`))
	})
}
