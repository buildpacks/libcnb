/*
 * Copyright 2018-2020, VMware, Inc. All Rights Reserved.
 * Proprietary and Confidential.
 * Unauthorized use, copying or distribution of this source code via any medium is
 * strictly prohibited without the express written consent of VMware, Inc.
 */

package libcnb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb/internal"
)

// Main is called by the main function of a buildpack, encapsulating both detection and build in the same binary.
func Main(detector Detector, builder Builder, options ...Option) {
	config := Config{
		arguments:         os.Args,
		environmentWriter: internal.EnvironmentWriter{},
		exitHandler:       internal.NewExitHandler(),
		tomlWriter:        internal.TOMLWriter{},
	}

	for _, option := range options {
		config = option(config)
	}

	if len(config.arguments) == 0 {
		config.exitHandler.Error(fmt.Errorf("expected command name"))
		return
	}

	switch c := filepath.Base(config.arguments[0]); c {
	case "build":
		Build(builder, options...)
	case "detect":
		Detect(detector, options...)
	default:
		config.exitHandler.Error(fmt.Errorf("unsupported command %s", c))
		return
	}
}
