/*
 * Copyright 2018-2021 the original author or authors.
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
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/stretchr/testify/mock"

	"github.com/buildpacks/libcnb"
	"github.com/buildpacks/libcnb/mocks"
)

func testExecD(t *testing.T, _ spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		exitHandler *mocks.ExitHandler
		execdWriter *mocks.ExecDWriter
	)

	it.Before(func() {
		execdWriter = &mocks.ExecDWriter{}
		execdWriter.On("Write", mock.Anything).Return(nil)
		exitHandler = &mocks.ExitHandler{}
		exitHandler.On("Error", mock.Anything)
		exitHandler.On("Pass", mock.Anything)
		exitHandler.On("Fail", mock.Anything)
	})

	it("encounters the wrong number of arguments", func() {
		libcnb.RunExecD(map[string]libcnb.ExecD{},
			libcnb.WithArguments([]string{}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("expected command name"))
	})

	it("encounters an unsupported execd binary name", func() {
		libcnb.RunExecD(map[string]libcnb.ExecD{},
			libcnb.WithArguments([]string{"/dne"}),
			libcnb.WithExitHandler(exitHandler),
		)

		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError("unsupported command dne"))
	})

	it("calls the appropriate execd for a given execd invoker binary", func() {
		execd1 := &mocks.ExecD{}
		execd2 := &mocks.ExecD{}
		execd1.On("Execute", mock.Anything).Return(map[string]string{}, nil)

		libcnb.RunExecD(map[string]libcnb.ExecD{"execd1": execd1, "execd2": execd2},
			libcnb.WithArguments([]string{"execd1"}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithExecDWriter(execdWriter),
		)

		Expect(execd1.Calls).To(HaveLen(1))
		Expect(execd2.Calls).To(BeEmpty())
	})

	it("calls exitHandler with the error from the execd", func() {
		e := &mocks.ExecD{}
		err := fmt.Errorf("example error")
		e.On("Execute", mock.Anything).Return(nil, err)

		libcnb.RunExecD(map[string]libcnb.ExecD{"e": e},
			libcnb.WithArguments([]string{"/bin/e"}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithExecDWriter(execdWriter),
		)

		Expect(e.Calls).To(HaveLen(1))
		Expect(execdWriter.Calls).To(HaveLen(0))
		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError(err))
	})

	it("calls execdWriter.write with the appropriate input", func() {
		e := &mocks.ExecD{}
		o := map[string]string{"test": "test"}
		e.On("Execute", mock.Anything).Return(o, nil)

		libcnb.RunExecD(map[string]libcnb.ExecD{"e": e},
			libcnb.WithArguments([]string{"/bin/e"}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithExecDWriter(execdWriter),
		)

		Expect(e.Calls).To(HaveLen(1))
		Expect(execdWriter.Calls).To(HaveLen(1))
		Expect(execdWriter.Calls[0].Method).To(BeIdenticalTo("Write"))
		Expect(execdWriter.Calls[0].Arguments).To(HaveLen(1))
		Expect(execdWriter.Calls[0].Arguments[0]).To(Equal(o))
	})

	it("calls exitHandler with the error from the execd", func() {
		e := &mocks.ExecD{}
		err := fmt.Errorf("example error")
		e.On("Execute", mock.Anything).Return(nil, err)

		libcnb.RunExecD(map[string]libcnb.ExecD{"e": e},
			libcnb.WithArguments([]string{"/bin/e"}),
			libcnb.WithExitHandler(exitHandler),
			libcnb.WithExecDWriter(execdWriter),
		)

		Expect(e.Calls).To(HaveLen(1))
		Expect(execdWriter.Calls).To(HaveLen(0))
		Expect(exitHandler.Calls[0].Arguments.Get(0)).To(MatchError(err))
	})
}
