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

package libcnb

import "github.com/buildpacks/libcnb/internal"

// Fail is a sentinal value that can be used to indicate a failure to detect
// during the detect phase. Fail implements the Error interface and should be
// returned as the error value in the DetectFunc signature. Fail also supports
// a modifier function, WithMessage, that allows the caller to set a custom
// failure message. The WithMessage function supports a fmt.Printf-like format
// string and variadic arguments to build a message, eg:
// packit.Fail.WithMessage("failed: %w", err).
var Fail = internal.Fail
