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

package internal

import (
	"fmt"
	"reflect"

	"github.com/BurntSushi/toml"
	"github.com/onsi/gomega/types"
)

func MatchTOML(expected interface{}) types.GomegaMatcher {
	return &matchTOML{
		expected: expected,
	}
}

type matchTOML struct {
	expected interface{}
}

func (matcher *matchTOML) Match(actual interface{}) (success bool, err error) {
	var e, a string

	switch eType := matcher.expected.(type) {
	case string:
		e = eType
	case []byte:
		e = string(eType)
	default:
		return false, fmt.Errorf("expected value must be []byte or string, received %T", matcher.expected)
	}

	switch aType := actual.(type) {
	case string:
		a = aType
	case []byte:
		a = string(aType)
	default:
		return false, fmt.Errorf("actual value must be []byte or string, received %T", matcher.expected)
	}

	var eValue map[string]interface{}
	_, err = toml.Decode(e, &eValue)
	if err != nil {
		return false, err
	}

	var aValue map[string]interface{}
	_, err = toml.Decode(a, &aValue)
	if err != nil {
		return false, err
	}

	return reflect.DeepEqual(eValue, aValue), nil
}

func (matcher *matchTOML) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nto match the TOML representation of\n%s", actual, matcher.expected)
}

func (matcher *matchTOML) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nnot to match the TOML representation of\n%s", actual, matcher.expected)
}
