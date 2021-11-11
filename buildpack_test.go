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
	"bytes"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/buildpacks/libcnb"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuildpackTOML(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("does not serialize the Path field", func() {
		bp := libcnb.Buildpack{
			API: "0.6",
			Info: libcnb.BuildpackInfo{
				ID:   "test-buildpack/sample",
				Name: "sample",
			},
			Path: "../buildpack",
		}

		output := &bytes.Buffer{}

		Expect(toml.NewEncoder(output).Encode(bp)).To(Succeed())
		Expect(output.String()).NotTo(ContainSubstring("ath = ")) // match on path and Path
	})
}
