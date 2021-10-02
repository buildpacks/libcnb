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
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb/internal"
)

func testExecDWriter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		b      *bytes.Buffer
		writer internal.ExecDWriter
	)

	it.Before(func() {
		b = bytes.NewBuffer([]byte{})

		writer = internal.NewExecDWriter(
			internal.WithExecDOutputWriter(b),
		)
	})

	it("writes the correct set of values", func() {
		env := map[string]string{
			"test":  "test",
			"test2": "te∆t",
		}
		Expect(writer.Write(env)).To(BeNil())
		Expect(b.String()).To(internal.MatchTOML(`
			test = "test"
			test2 = "te∆t"`))
	})
}
