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
	"sort"

	"github.com/buildpacks/libcnb/internal"
)

const (
	// BindingKind is the metadata key for a binding's kind.
	BindingKind = "kind"

	// BindingProvider is the key for a binding's provider.
	BindingProvider = "provider"

	// BindingType is the key for a binding's type.
	BindingType = "type"
)

// Binding is a projection of metadata about an external entity to be bound to.
type Binding struct {

	// Name is the name of the binding
	Name string

	// Path is the path to the binding directory.
	Path string

	// Type is the type of the binding.
	Type string

	// Provider is the optional provider of the binding.
	Provider string

	// Secret is the secret of the binding.
	Secret map[string]string
}

// NewBinding creates a new Binding initialized with a secret.
func NewBinding(name string, path string, secret map[string]string) Binding {
	b := Binding{
		Name:   name,
		Path:   path,
		Secret: make(map[string]string),
	}

	for k, v := range secret {
		if k == BindingType {
			b.Type = v
		} else if k == BindingProvider {
			b.Provider = v
		} else if k == BindingKind { // TODO: Remove as CNB_BINDINGS ages out
			b.Type = v
		} else {
			b.Secret[k] = v
		}
	}

	return b
}

// NewBindingFromPath creates a new binding from the files located at a path.
func NewBindingFromPath(path string) (Binding, error) {
	secret, err := internal.NewConfigMapFromPath(path)
	if err != nil {
		return Binding{}, fmt.Errorf("unable to create new config map from %s\n%w", path, err)
	}

	// TODO: Remove as CNB_BINDINGS ages out
	for _, f := range []string{filepath.Join(path, "metadata"), filepath.Join(path, "secret")} {
		cm, err := internal.NewConfigMapFromPath(f)
		if err != nil {
			return Binding{}, fmt.Errorf("unable to create new config map from %s\n%w", f, err)
		}

		for k, v := range cm {
			secret[k] = v
		}
	}

	return NewBinding(filepath.Base(path), path, secret), nil
}

func (b Binding) String() string {
	var s []string
	for k, _ := range b.Secret {
		s = append(s, k)
	}
	sort.Strings(s)

	return fmt.Sprintf("{Path: %s Secret: %s}", b.Path, s)
}

// SecretFilePath return the path to a secret file with the given name.
func (b Binding) SecretFilePath(name string) (string, bool) {
	if _, ok := b.Secret[name]; !ok {
		return "", false
	}

	// TODO: Remove as CNB_BINDINGS ages out
	for _, d := range []string{"metadata", "secret"} {
		f := filepath.Join(b.Path, d, name)
		if _, err := os.Stat(f); err == nil {
			return f, true
		}
	}

	return filepath.Join(b.Path, name), true
}

// Bindings is a collection of bindings keyed by their name.
type Bindings []Binding

// NewBindingsFromEnvironment creates a new bindings from all the bindings at the path defined by $SERVICE_BINDING_ROOT
// or $CNB_BINDINGS if it does not exist.  If neither is defined, returns an empty collection of Bindings.
func NewBindingsFromEnvironment() (Bindings, error) {
	if path, ok := os.LookupEnv("SERVICE_BINDING_ROOT"); ok {
		return NewBindingsFromPath(path)
	}

	// TODO: Remove as CNB_BINDINGS ages out
	if path, ok := os.LookupEnv("CNB_BINDINGS"); ok {
		return NewBindingsFromPath(path)
	}

	return Bindings{}, nil
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

		bindings = append(bindings, binding)
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
