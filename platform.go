/*
 * Copyright 2018-2023 the original author or authors.
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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/buildpacks/libcnb/internal"
)

const (
	// BindingProvider is the key for a binding's provider.
	BindingProvider = "provider"

	// BindingType is the key for a binding's type.
	BindingType = "type"

	// EnvServiceBindings is the name of the environment variable that contains the path to service bindings directory.
	//
	// See the Service Binding Specification for Kubernetes for more details - https://k8s-service-bindings.github.io/spec/
	EnvServiceBindings = "SERVICE_BINDING_ROOT"

	// EnvBuildpackDirectory is the name of the environment variable that contains the path to the buildpack
	EnvBuildpackDirectory = "CNB_BUILDPACK_DIR"

	// EnvExtensionDirectory is the name of the environment variable that contains the path to the extension
	EnvExtensionDirectory = "CNB_EXTENSION_DIR"

	// EnvVcapServices is the name of the environment variable that contains the bindings in cloudfoundry
	EnvVcapServices = "VCAP_SERVICES"

	// EnvLayersDirectory is the name of the environment variable that contains the root path to all buildpack layers
	EnvLayersDirectory = "CNB_LAYERS_DIR"

	// EnvOutputDirectory is the name of the environment variable that contains the path to the output directory
	EnvOutputDirectory = "CNB_OUTPUT_DIR"

	// EnvPlatformDirectory is the name of the environment variable that contains the path to the platform directory
	EnvPlatformDirectory = "CNB_PLATFORM_DIR"

	// EnvDetectBuildPlanPath is the name of the environment variable that contains the path to the build plan
	EnvDetectPlanPath = "CNB_BUILD_PLAN_PATH"

	// EnvBuildPlanPath is the name of the environment variable that contains the path to the build plan
	EnvBuildPlanPath = "CNB_BP_PLAN_PATH"

	// EnvStackID is the name of the environment variable that contains the stack id
	EnvStackID = "CNB_STACK_ID"

	// DefaultPlatformBindingsLocation is the typical location for bindings, which exists under the platform directory
	//
	// Not guaranteed to exist, but often does. This should only be used as a fallback if EnvServiceBindings and EnvPlatformDirectory are not set
	DefaultPlatformBindingsLocation = "/platform/bindings"
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
		case BindingType:
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

	return NewBinding(filepath.Base(path), path, secret), nil
}

func (b Binding) String() string {
	var s []string
	for k := range b.Secret {
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

	return filepath.Join(b.Path, name), true
}

// Bindings is a collection of bindings keyed by their name.
type Bindings []Binding

// NewBindingsFromPath creates a new instance from all the bindings at a given path.
func NewBindingsFromPath(path string) (Bindings, error) {
	files, err := os.ReadDir(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		return Bindings{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to list directory %s\n%w", path, err)
	}

	bindings := Bindings{}
	for _, file := range files {
		bindingPath := filepath.Join(path, file.Name())

		if strings.HasPrefix(filepath.Base(bindingPath), ".") {
			// ignore hidden files
			continue
		}
		binding, err := NewBindingFromPath(bindingPath)
		if err != nil {
			return nil, fmt.Errorf("unable to create new binding from %s\n%w", file, err)
		}

		bindings = append(bindings, binding)
	}

	return bindings, nil
}

type vcapServicesBinding struct {
	Name        string            `json:"name"`
	Label       string            `json:"label"`
	Credentials map[string]string `json:"credentials"`
}

// NewBindingsFromVcapServicesEnv creates a new instance from all the bindings given from the VCAP_SERVICES.
func NewBindingsFromVcapServicesEnv(content string) (Bindings, error) {
	var contentTyped map[string][]vcapServicesBinding

	err := json.Unmarshal([]byte(content), &contentTyped)
	if err != nil {
		return Bindings{}, nil
	}

	bindings := Bindings{}
	for p, bArray := range contentTyped {
		for _, b := range bArray {
			bindings = append(bindings, Binding{
				Name:     b.Name,
				Type:     b.Label,
				Provider: p,
				Secret:   b.Credentials,
			})
		}
	}

	return bindings, nil
}

// NewBindings creates a new bindings from all the bindings at the path defined by $SERVICE_BINDING_ROOT.
// If that isn't defined, bindings are read from <platform>/bindings.
// If that isn't defined, bindings are read from $VCAP_SERVICES.
// If that isn't defined, the specified platform path will be used
func NewBindings(platformDir string) (Bindings, error) {
	if path, ok := os.LookupEnv(EnvServiceBindings); ok {
		return NewBindingsFromPath(path)
	}

	if path, ok := os.LookupEnv(EnvPlatformDirectory); ok {
		return NewBindingsFromPath(filepath.Join(path, "bindings"))
	}

	if content, ok := os.LookupEnv(EnvVcapServices); ok {
		return NewBindingsFromVcapServicesEnv(content)
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
