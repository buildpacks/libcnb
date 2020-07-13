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

	// BindingProvider is the metadata key for a binding's provider.
	BindingProvider = "provider"

	// BindingType is the metadata key for a binding's type.
	BindingType = "type"
)

// Binding is a projection of metadata about an external entity to be bound to.
type Binding struct {

	// Name is the name of the binding
	Name string

	// Secret is the secret of the binding.
	Secret map[string]string

	// Path is the path to the binding directory.
	Path string
}

// NewBinding creates a new Binding initialized with no metadata or secret.
func NewBinding(name string) Binding {
	return Binding{
		Name:   name,
		Secret: map[string]string{},
	}
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

	return Binding{
		Name:   filepath.Base(path),
		Path:   path,
		Secret: secret,
	}, nil
}

// Type returns the type of the binding.
func (b Binding) Type() string {
	if s, ok := b.Secret[BindingType]; ok {
		return s
	}

	return b.Secret[BindingKind]
}

// Provider returns the provider of the binding.
func (b Binding) Provider() string {
	return b.Secret[BindingProvider]
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
