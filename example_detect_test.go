package libcnb_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/buildpacks/libcnb/log"
)

const (
	Provides         = "example"
	BpExampleVersion = "BP_EXAMPLE_VERSION"
)

type Detector struct {
	Logger libcnb.Logger
}

func (Detector) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	version := "1.0"
	// Scan the application source folder to see if the example buildpack is
	// required.  If `version.toml` does not exist we return a failed DetectResult
	// but no runtime error has occurred, so we return an empty error.
	versionPath := filepath.Join(context.ApplicationPath, "version.toml")
	if _, err := os.Open(versionPath); errors.Is(err, os.ErrNotExist) {
		return libcnb.DetectResult{}, nil
	}
	// Read the version number from the buildpack definition
	if exampleVersion, exists := context.Buildpack.Metadata["version"]; exists {
		version = exampleVersion.(string)
	}
	// Accept version number from the environment if the user provides it
	if exampleVersion, exists := context.Platform.Environment[BpExampleVersion]; exists {
		version = exampleVersion
	}
	metadata := map[string]interface{}{
		"version": version,
	}
	return libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				// Let the system know that if other buildpacks Require "example"
				// then this buildpack Provides the implementation logic.
				Provides: []libcnb.BuildPlanProvide{
					{Name: Provides},
				},
				// It is common for a buildpack to Require itself if the build phase
				// needs information from the detect phase. Here we pass the version number
				// as metadata to the build phase.
				Requires: []libcnb.BuildPlanRequire{
					{
						Name:     Provides,
						Metadata: metadata,
					},
				},
			},
		},
	}, nil
}

func ExampleDetect() {
	detector := Detector{log.New(os.Stdout)}
	libcnb.Main(detector.Detect, nil)
}
