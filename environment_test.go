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
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb"
)

func testEnvironment(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		environment libcnb.Environment
	)

	it.Before(func() {
		environment = libcnb.Environment{}
	})

	it("adds append value", func() {
		environment.Append("TEST_NAME", "test-delimiter", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{
			"TEST_NAME.delim":  "test-delimiter",
			"TEST_NAME.append": "test-value",
		}))
	})

	it("adds append formatted value", func() {
		environment.Appendf("TEST_NAME", "test-delimiter", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{
			"TEST_NAME.delim":  "test-delimiter",
			"TEST_NAME.append": "test-value",
		}))
	})

	it("adds default value", func() {
		environment.Default("TEST_NAME", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{"TEST_NAME.default": "test-value"}))
	})

	it("adds default formatted value", func() {
		environment.Defaultf("TEST_NAME", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{"TEST_NAME.default": "test-value"}))
	})

	it("adds override value", func() {
		environment.Override("TEST_NAME", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{"TEST_NAME.override": "test-value"}))
	})

	it("adds override formatted value", func() {
		environment.Overridef("TEST_NAME", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{"TEST_NAME.override": "test-value"}))
	})

	it("adds prepend value", func() {
		environment.Prepend("TEST_NAME", "test-delimiter", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{
			"TEST_NAME.delim":   "test-delimiter",
			"TEST_NAME.prepend": "test-value",
		}))
	})

	it("adds prepend formatted value", func() {
		environment.Prependf("TEST_NAME", "test-delimiter", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{
			"TEST_NAME.delim":   "test-delimiter",
			"TEST_NAME.prepend": "test-value",
		}))
	})

	it("adds process-specific append value", func() {
		environment.ProcessAppend("test-process", "TEST_NAME", "test-delimiter", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{
			filepath.Join("test-process", "TEST_NAME.delim"):  "test-delimiter",
			filepath.Join("test-process", "TEST_NAME.append"): "test-value",
		}))
	})

	it("adds process-specific append formatted value", func() {
		environment.ProcessAppendf("test-process", "TEST_NAME", "test-delimiter", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{
			filepath.Join("test-process", "TEST_NAME.delim"):  "test-delimiter",
			filepath.Join("test-process", "TEST_NAME.append"): "test-value",
		}))
	})

	it("adds process-specific default value", func() {
		environment.ProcessDefault("test-process", "TEST_NAME", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{filepath.Join("test-process", "TEST_NAME.default"): "test-value"}))
	})

	it("adds process-specific default formatted value", func() {
		environment.ProcessDefaultf("test-process", "TEST_NAME", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{filepath.Join("test-process", "TEST_NAME.default"): "test-value"}))
	})

	it("adds process-specific override value", func() {
		environment.ProcessOverride("test-process", "TEST_NAME", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{filepath.Join("test-process", "TEST_NAME.override"): "test-value"}))
	})

	it("adds process-specific override formatted value", func() {
		environment.ProcessOverridef("test-process", "TEST_NAME", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{filepath.Join("test-process", "TEST_NAME.override"): "test-value"}))
	})

	it("adds process-specific prepend value", func() {
		environment.ProcessPrepend("test-process", "TEST_NAME", "test-delimiter", "test-value")
		Expect(environment).To(Equal(libcnb.Environment{
			filepath.Join("test-process", "TEST_NAME.delim"):   "test-delimiter",
			filepath.Join("test-process", "TEST_NAME.prepend"): "test-value",
		}))
	})

	it("adds process-specific prepend formatted value", func() {
		environment.ProcessPrependf("test-process", "TEST_NAME", "test-delimiter", "test-%s", "value")
		Expect(environment).To(Equal(libcnb.Environment{
			filepath.Join("test-process", "TEST_NAME.delim"):   "test-delimiter",
			filepath.Join("test-process", "TEST_NAME.prepend"): "test-value",
		}))
	})

}
