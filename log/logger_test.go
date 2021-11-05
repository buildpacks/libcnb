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

package log_test

import (
	"bytes"
	"os"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/buildpacks/libcnb/log"
)

func testLogger(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		b *bytes.Buffer
		l log.Logger
	)

	it.Before(func() {
		b = bytes.NewBuffer(nil)
	})

	context("without BP_DEBUG", func() {
		it.Before(func() {
			l = log.New(b)
		})

		it("does not configure debug", func() {
			Expect(l.IsDebugEnabled()).To(BeFalse())
		})
	})

	context("with BP_DEBUG", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_DEBUG", "")).To(Succeed())
			l = log.New(b)
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_DEBUG")).To(Succeed())
		})

		it("configures debug", func() {
			Expect(l.IsDebugEnabled()).To(BeTrue())
		})
	})

	context("with BP_LOG_LEVEL set to DEBUG", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_LOG_LEVEL", "DEBUG")).To(Succeed())
			l = log.New(b)
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_LOG_LEVEL")).To(Succeed())
		})

		it("configures debug", func() {
			Expect(l.IsDebugEnabled()).To(BeTrue())
		})
	})

	context("with debug disabled", func() {
		it.Before(func() {
			l = log.NewWithOptions(b)
		})

		it("does not write debug log", func() {
			l.Debug("test-message")
			Expect(b.String()).To(Equal(""))
		})

		it("does not write debugf log", func() {
			l.Debugf("test-%s", "message")
			Expect(b.String()).To(Equal(""))
		})

		it("does not return debug writer", func() {
			Expect(l.DebugWriter()).To(BeNil())
		})

		it("indicates that debug is not enabled", func() {
			Expect(l.IsDebugEnabled()).To(BeFalse())
		})

		it("writes info log", func() {
			l.Info("test-message")
			Expect(b.String()).To(Equal("test-message\n"))
		})

		it("writes infof log", func() {
			l.Infof("test-%s", "message")
			Expect(b.String()).To(Equal("test-message\n"))
		})

		it("returns info writer", func() {
			Expect(l.InfoWriter()).NotTo(BeNil())
		})

		it("indicates that info is enabled", func() {
			Expect(l.IsInfoEnabled()).To(BeTrue())
		})
	})

	context("with debug enabled", func() {
		it.Before(func() {
			l = log.NewWithOptions(b, log.WithDebug(b))
		})

		it("writes debug log", func() {
			l.Debug("test-message")
			Expect(b.String()).To(Equal("test-message\n"))
		})

		it("writes debugf log", func() {
			l.Debugf("test-%s", "message")
			Expect(b.String()).To(Equal("test-message\n"))
		})

		it("returns debug writer", func() {
			Expect(l.DebugWriter()).NotTo(BeNil())
		})

		it("indicates that debug is enabled", func() {
			Expect(l.IsDebugEnabled()).To(BeTrue())
		})

		it("writes info log", func() {
			l.Info("test-message")
			Expect(b.String()).To(Equal("test-message\n"))
		})

		it("writes infof log", func() {
			l.Infof("test-%s", "message")
			Expect(b.String()).To(Equal("test-message\n"))
		})

		it("returns info writer", func() {
			Expect(l.InfoWriter()).NotTo(BeNil())
		})

		it("indicates that info is enabled", func() {
			Expect(l.IsInfoEnabled()).To(BeTrue())
		})
	})
}
