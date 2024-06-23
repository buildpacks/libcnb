package examples

import (
	"fmt"
	"os"

	"github.com/buildpacks/libcnb/v2"
	"github.com/buildpacks/libcnb/v2/log"
)

type Generator struct {
	Logger log.Logger
}

func (Generator) Generate(context libcnb.GenerateContext) (libcnb.GenerateResult, error) {
	// here you can read the context.ApplicationPath folder
	// and create run.Dockerfile and build.Dockerfile in the context.OutputPath folder
	// and read metadata from the context.Extension struct

	// Just to use context to keep compiler happy =)
	fmt.Println(context.Extension.Info.ID)

	result := libcnb.NewGenerateResult()
	return result, nil
}

func ExampleGenerate() {
	generator := Generator{log.New(os.Stdout)}
	libcnb.ExtensionMain(nil, generator.Generate)
}
