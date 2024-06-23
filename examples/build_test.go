package examples

import (
	"fmt"
	"os"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"

	"github.com/buildpacks/libcnb/v2"
	"github.com/buildpacks/libcnb/v2/log"
)

const (
	DefaultVersion = "0.1"
)

type Builder struct {
	Logger log.Logger
}

// BuildpackPlan may contain multiple entries for a single buildpack, resolve
// into a single entry.
func resolve(plan libcnb.BuildpackPlan[string], name string) libcnb.BuildpackPlanEntry[string] {
	entry := libcnb.BuildpackPlanEntry[string]{
		Name:     name,
		Metadata: map[string]string{},
	}
	for _, e := range plan.Entries {
		for k, v := range e.Metadata {
			entry.Metadata[k] = v
		}
	}
	return entry
}

func populateLayer(layer libcnb.Layer[string], version string) (libcnb.Layer[string], error) {
	exampleFile := filepath.Join(layer.Path, "example.txt")
	if err := os.WriteFile(exampleFile, []byte(version), 0600); err != nil {
		return libcnb.Layer[string]{}, fmt.Errorf("unable to write example file: %w", err)
	}

	layer.SharedEnvironment.Default("EXAMPLE_FILE", exampleFile)

	// Provide an SBOM
	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component: &cdx.Component{
			Type:    cdx.ComponentTypeFile,
			Name:    "example",
			Version: version,
		},
	}
	sbomPath := layer.SBOMPath(libcnb.CycloneDXJSON)
	sbomFile, err := os.OpenFile(sbomPath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return layer, err
	}
	defer sbomFile.Close()
	encoder := cdx.NewBOMEncoder(sbomFile, cdx.BOMFileFormatJSON)
	if err := encoder.Encode(bom); err != nil {
		return layer, err
	}
	return layer, nil
}

func (b Builder) Build(context libcnb.BuildContext[string, string, string, string]) (libcnb.BuildResult[string, string], error) {
	// Reduce possible multiple buildpack plan entries to a single entry
	entry := resolve(context.Plan, Provides)
	result := libcnb.NewBuildResult[string, string]()

	// Read metadata from the buildpack plan, often contributed by libcnb.Requires
	// of the Detect phase
	version := DefaultVersion
	if v, ok := entry.Metadata["version"]; ok {
		version = v
	}

	// Create a layer
	layer, err := context.Layers.Layer("example")
	if err != nil {
		return result, err
	}
	layer.LayerTypes = libcnb.LayerTypes{
		Launch: true,
		Build:  true,
		Cache:  true,
	}

	layer, err = populateLayer(layer, version)
	if err != nil {
		return result, nil
	}

	result.Layers = append(result.Layers, layer)
	return result, nil
}

func ExampleBuild() {
	detector := Detector{log.New(os.Stdout)}
	builder := Builder{log.New(os.Stdout)}
	libcnb.BuildpackMain(detector.Detect, builder.Build)
}
