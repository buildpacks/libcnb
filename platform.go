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
	"strings"

	"github.com/buildpacks/libcnb/internal"
)

const (
	// BindingKind is the metadata key for a binding's kind.
	BindingKind = "kind"

	// BindingProvider is the key for a binding's provider.
	BindingProvider = "provider"

	// BindingType is the key for a binding's type.
	BindingType = "type"

	// EnvServiceBindings is the name of the environment variable that contains the path to service bindings directory.
	//
	// See the Service Binding Specification for Kubernetes for more details - https://k8s-service-bindings.github.io/spec/
	EnvServiceBindings = "SERVICE_BINDING_ROOT"

	// EnvCNBBindings is the name of the environment variable that contains the path to the CNB bindings directory. The CNB
	// bindings spec will eventually by deprecated in favor of the Service Binding Specification for Kubernetes -
	// https://github.com/buildpacks/rfcs/blob/main/text/0055-deprecate-service-bindings.md.
	//
	// See the CNB bindings extension spec for more details - https://github.com/buildpacks/spec/blob/main/extensions/bindings.md
	EnvCNBBindings = "CNB_BINDINGS"
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
		switch k {
		case BindingType, BindingKind: // TODO: Remove as CNB_BINDINGS ages out
			b.Type = strings.TrimSpace(v)
		case BindingProvider:
			b.Provider = strings.TrimSpace(v)
		default:
			b.Secret[k] = strings.TrimSpace(v)
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
	for _, d := range []string{"metadata", "secret"} {
		file := filepath.Join(path, d)
		cm, err := internal.NewConfigMapFromPath(file)
		if err != nil {
			return Binding{}, fmt.Errorf("unable to create new config map from %s\n%w", file, err)
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

	return fmt.Sprintf("{Name: %s Path: %s Type: %s Provider: %s Secret: %s}",
		b.Name, b.Path, b.Type, b.Provider, s)
}

// SecretFilePath return the path to a secret file with the given name.
func (b Binding) SecretFilePath(name string) (string, bool) {
	if _, ok := b.Secret[name]; !ok {
		return "", false
	}

	// TODO: Remove as CNB_BINDINGS ages out
	for _, d := range []string{"metadata", "secret"} {
		file := filepath.Join(b.Path, d, name)
		if _, err := os.Stat(file); err == nil {
			return file, true
		}
	}

	return filepath.Join(b.Path, name), true
}

// Bindings is a collection of bindings keyed by their name.
type Bindings []Binding

// NewBindingsFromEnvironment creates a new bindings from all the bindings at the path defined by $SERVICE_BINDING_ROOT
// or $CNB_BINDINGS if it does not exist.  If neither is defined, returns an empty collection of Bindings.
// Note - This API is deprecated. Please use NewBindingsForLaunch instead.
func NewBindingsFromEnvironment() (Bindings, error) {
	return NewBindingsForLaunch()
}

// NewBindingsForLaunch creates a new bindings from all the bindings at the path defined by $SERVICE_BINDING_ROOT
// or $CNB_BINDINGS if it does not exist.  If neither is defined, returns an empty collection of Bindings.
func NewBindingsForLaunch() (Bindings, error) {
	if path, ok := os.LookupEnv(EnvServiceBindings); ok {
		return NewBindingsFromPath(path)
	}

	// TODO: Remove as CNB_BINDINGS ages out
	if path, ok := os.LookupEnv(EnvCNBBindings); ok {
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

// NewBindingsForBuild creates a new bindings from all the bindings at the path defined by $SERVICE_BINDING_ROOT
// or $CNB_BINDINGS if it does not exist.  If neither is defined, bindings are read from <platform>/bindings, the default
// path defined in the CNB Binding extension specification.
func NewBindingsForBuild(platformDir string) (Bindings, error) {
	if path, ok := os.LookupEnv(EnvServiceBindings); ok {
		return NewBindingsFromPath(path)
	}
	// TODO: Remove as CNB_BINDINGS ages out
	if path, ok := os.LookupEnv(EnvCNBBindings); ok {
		return NewBindingsFromPath(path)
	}
	return NewBindingsFromPath(filepath.Join(platformDir, "bindings"))
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
