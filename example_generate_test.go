package libcnb_test

import (
	"os"

	"github.com/buildpacks/libcnb"
	"github.com/buildpacks/libcnb/log"
)

type Generator struct {
	Logger log.Logger
}

func (Generator) Generate(context libcnb.GenerateContext) (libcnb.GenerateResult, error) {
	//here you can read the context.ApplicationPath folder
	//and create run.Dockerfile and build.Dockerfile in the context.OutputPath folder
	//and read metadata from the context.Extension struct
	result := libcnb.NewGenerateResult()
	return result, nil
}

func ExampleGeneratre() {
	generator := Generator{log.New(os.Stdout)}
	libcnb.ExtensionMain(nil, generator.Generate)
}
