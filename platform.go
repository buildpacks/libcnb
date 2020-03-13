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

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/internal"
)

const (

	// BindingKind is the metadata key for a binding's kind.
	BindingKind = "kind"

	// BindingProvider is the metadata key for a binding's provider.
	BindingProvider = "provider"

	// BindingTags is the metadata key for a binding's tags.
	BindingTags = "tags"
)

// Binding is a projection of metadata about an external entity to be bound to.
type Binding struct {

	// Metadata is the metadata of the binding.
	Metadata map[string]string `toml:"metadata"`

	// Secrete is the secret of the binding.
	Secret map[string]string `toml:"secret"`
}

// NewBinding creates a new Binding initialized with no metadata or secret.
func NewBinding() Binding {
	return Binding{
		Metadata: map[string]string{},
		Secret:   map[string]string{},
	}
}

// NewBindingFromPath creates a new binding from the files located at a path.
func NewBindingFromPath(path string) (Binding, error) {
	var f string

	f = filepath.Join(path, "metadata")
	metadata, err := internal.NewConfigMapFromPath(f)
	if err != nil {
		return Binding{}, fmt.Errorf("unable to create new config map from %s\n%w", f, err)
	}

	f = filepath.Join(path, "secret")
	secret, err := internal.NewConfigMapFromPath(f)
	if err != nil {
		return Binding{}, fmt.Errorf("unable to create new config map from %s\n%w", f, err)
	}

	return Binding{metadata, secret}, nil
}

func (b Binding) String() string {
	var s []string
	for k, _ := range b.Secret {
		s = append(s, k)
	}

	return fmt.Sprintf("{Metadata: %s Secret: %s}", b.Metadata, s)
}

// Bindings is a collection of bindings keyed by their name.
type Bindings map[string]Binding

// NewBindingsFromEnvironment creates a new bindings from all the bindings at the path defined by $CNB_BINDINGS.  If
// $CNB_BINDINGS is not defined, returns an empty collection of Bindings.
func NewBindingsFromEnvironment() (Bindings, error) {
	path, ok := os.LookupEnv("CNB_BINDINGS")
	if !ok {
		return Bindings{}, nil
	}

	return NewBindingsFromPath(path)
}

// NewBindingsFromPath creates a new instance from all the bindings at a given path.
func NewBindingsFromPath(path string) (Bindings, error) {
	files, err := filepath.Glob(filepath.Join(path, "*"))
	if err != nil {
		return nil, fmt.Errorf("unable to glob %s\n%w", path, err)
	}

	bindings := Bindings{}
	for _, file := range files {
		binding, err := NewBindingFromPath(file)
		if err != nil {
			return nil, fmt.Errorf("unable to create new binding from %s\n%w", file, err)
		}

		name := filepath.Base(file)

		bindings[name] = binding
	}

	return bindings, nil
}

// Platform is the contents of the platform directory.
type Platform struct {

	// Bindings are the external bindings available to the application.
	Bindings Bindings

	// Environment is the environment exposed by the platform.
	Environment map[string]string

	// Path is the path to the platform.
	Path string
}
