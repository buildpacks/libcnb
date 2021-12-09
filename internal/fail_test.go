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
	"testing"

	"github.com/buildpacks/libcnb/internal"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testFail(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	it("acts as an error", func() {
		fail := internal.Fail
		Expect(fail).To(MatchError("failed"))
	})

	context("when given a message", func() {
		it("acts as an error with that message", func() {
			fail := internal.Fail.WithMessage("this is a %s", "failure message")
			Expect(fail).To(MatchError("this is a failure message"))
		})
	})
}
