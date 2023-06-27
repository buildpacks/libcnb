/*
 * Copyright 2023 the original author or authors.
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
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb"

	. "github.com/onsi/gomega"
)

func testExtensionTOML(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("does not serialize the Path field", func() {
		extn := libcnb.Extension{
			API: "0.8",
			Info: libcnb.ExtensionInfo{
				ID:   "test-buildpack/sample",
				Name: "sample",
			},
			Path: "../buildpack",
		}

		output := &bytes.Buffer{}

		Expect(toml.NewEncoder(output).Encode(extn)).To(Succeed())
		Expect(output.String()).NotTo(Or(ContainSubstring("Path = "), ContainSubstring("path = ")))
	})
}
