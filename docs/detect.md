# Detection

A buildpack is suitably packaged `detect` binary, a `build` binary and a `buildpack.toml`.  Here we look at the implementation logic of the detect binary for our example buildpack.  This starts by declaring a `struct` that we call `Detector`.  The purpose of our struct is to satisfy `libcnb`s [`Detector`](https://pkg.go.dev/github.com/buildpacks/libcnb?utm_source=gopls#Detector) interface providing a receiver with signature `func Detect(context DetectContext) (DetectResult, error)`.  An implementation of `Detector` is passed to `libcnb.Main` as the main entry point to the `detect` binary.

In our `Detector` we use a [`poet`](https://pkg.go.dev/github.com/buildpacks/libcnb/poet) logger instance.  The logger instance will be used to inform the user of the progress of our `detect` phase.

```go
type Detector struct {
	Logger poet.Logger
}
```

## Implementing Detect

The most simple implementation of `func (context DetectContext) (DetectResult, error)` returns a `DetectResult` ensuring that the this buildpack does not contribute to the build phase.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	return DetectResult{}, nil
}
```

The trivial passing implementation of `Detect` ensures that the `build` binary of this buildpack will run during the build phase.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	return DetectResult{Pass: true}, nil
}
```

These two trivial cases ignore the `Plans` field of the `DetectResult` struct.  In order to understand the `Plans` field we look at the concept of provides and requires.

### Provides and Requires

Buildpacks may depend on other buildpacks.  The dependencies are computed at runtime.  Buildpacks can **provide** functionality that other buildpacks use.  For example, a buildpack may provide a runtime such as the python runtime or a Java runtime environment; other buildpacks may detect that an application is a python or Java application and require the provided runtime.  Buildpacks may also **require** functionality provided by other buildpacks.

A buildpack can offer multiple provides.  However, it is common for a buildpack to offer only a single provide.  Each provision is identified by a string.  Here we declare `Example` as an identifier.

```go
const (
    Example = "example"
)
```

Extending a trivial passing implementation with `Provides` allows another buildpack to require this buildpack.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	return DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				// Let the system know that if other buildpacks Require "example"
				// then this buildpack Provides the implementation logic
				Provides: []libcnb.BuildPlanProvide{
					{Name: Example},
				},
			}
		}
	}, nil
}
```

The detect phase of another buildpack may require this buildpack.  This is commonly used when the detect phase gathers information that is passed as metadata to the build phase.  Similarly, it is common for a buildpack to require itself.  This will allow the detect phase of a buildpack to pass information to the build phase.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	return DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				// Let the system know that if other buildpacks Require "example"
				// then this buildpack Provides the implementation logic
				Provides: []libcnb.BuildPlanProvide{
					{Name: Example},
				},
				Requires: []libcmb.BuildPlanRequire{
					{Name: Example},
				}
			}
		}
	}, nil
}
```

Here we pass the key-value pair `(version, 1.0)` as metadata from the detect phase of the example buildpack to the build phase. Metadata can be arbitrarily nested, but it must be textual as it is serialized to TOML at the end of the detect phase.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	return DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				// Let the system know that if other buildpacks Require "example"
				// then this buildpack Provides the implementation logic
				Provides: []libcnb.BuildPlanProvide{
					{Name: Example},
				},
				Requires: []libcmb.BuildPlanRequire{
					{
						Name: Example,
						// Pass arbitrary metadata to the Example build phase
						Metadata: {
							"version": "1.0",
						}
					},
				}
			}
		}
	}, nil
}
```

The `detect` binary is provided with read-only access to the application source code.  As an example we can detect the presence of a file called `example.txt`.  If this file does not exist, then we can fail at the detect phase.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	const exampleFile := "example.txt"
	// Construct a path to exampleFile in the Application source project
	requirementsFile := filepath.Join(context.Application.Path, exampleFile)
	// Test for the existance of exampleFile
	if _, err := os.Stat(exampleFile); err != nil && os.IsNotExist(err) {
		return libcnb.DetectResult{}, err
	}
	d.Logger.Debugf("Found %s example file -> %s", exampleFile)
	...
}
```

We have defined a `detect` binary that passes if `example.txt` is present in the source application and fails otherwise.

During the detect phase the buildpack has access to static metadata declared in the `buildpack.toml` file.

## Accessing buildpack metadata

It is common for static metadata to be declared in a `buildpack.toml`.  This is commonly used to pass URLs to the buildpack that are used to download artefacts.  As an example, a `buildpack.toml` may declare the download location for the `example` artefact:

```toml
[metadata]
  [[metadata.dependencies]]
    url = "https://source.fake/releases/v{{ .version }}/example_{{ .version }}_linux_amd64.tar.gz"
```

In our running example the detect phase of our buildpack detects the presence of `example.txt`.  We may want a buildpack definition to change the name of the file for detection.

```toml
[metadata]
  [[metadata.dependencies]]
    example-file = "custom-example.txt"
```

The `Detect(context)` function can access the metadata in the buildpack definition.

```go
func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	var exampleFile := "example.txt"
	if configuredExampleFile, ok; context.Buildpack.Metadata["example-file"]; ok {
		exampleFile = configuredExampleFile.(string)
	}
	// Construct a path to exampleFile in the Application source project
	requirementsFile := filepath.Join(context.Application.Path, exampleFile)
	// Test for the existance of exampleFile
	if _, err := os.Stat(exampleFile); err != nil && os.IsNotExist(err) {
		return libcnb.DetectResult{Pass: true}, err
	}
	d.Logger.Debugf("Found %s example file -> %s", exampleFile)
	...
}
```

If `example.txt` can be overridden it is common to also provide an environment variable to allow overriding.  Here we allow `example.txt` to be overridden using either a static definition in the `buildpack.toml` file or by the presence of a `BP_EXAMPLE_FILE` environment variable.

```go
const (
    BpExampleFile = "BP_EXAMPLE_FILE"
)

func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	var exampleFile := "example.txt"
	if configuredExampleFile, ok; context.Buildpack.Metadata["example-file"]; ok {
		exampleFile = configuredExampleFile.(string)
	}
	if file, ok := context.Platform.Environment(BpExamplefile); ok {
	    exampleFile = file
	}
	// Construct a path to exampleFile in the Application source project
	requirementsFile := filepath.Join(context.Application.Path, exampleFile)
	// Test for the existance of exampleFile
	if _, err := os.Stat(exampleFile); err != nil && os.IsNotExist(err) {
		return libcnb.DetectResult{Pass: true}, err
	}
	d.Logger.Debugf("Found %s example file -> %s", exampleFile)
	...
}
```

Buildpack metadata and environment variables are useful when passing configuration to a buildpack `detect` binary.  Passing sensitive data, such as secrets, requires a different approach.

## Secrets and other bindings

Secrets are available through the `DetectContext`.  Here we look for a binding named "GIT_TOKEN" which we expect to define a secret called "TOKEN".  We assume the existence of a function with signature `findBinding([]libcnb.Bindings, string) (libcnb.Binding, error)`.

```go
const (
    GitTokenBinding = "GIT_TOKEN"
    GitToken = "TOKEN"
)

func (d Detector) Detect(context DetectContext) (DetectResult, error) {
	bindings := context.Platform.Bindings
	gitTokenBinding, err := findBinding(bindings, GitTokenBinding)
	if err != nil {
		return DetectResult{}, fmt.errorf("unable to access binding %s", GitTokenBinding)
	}
	if gitToken, exists := gitTokenBinding.Secret[GitToken]; !exists {
		return DetectResult{}, fmt.errorf("unable to find secret %s in binding %s", GitToken, GitTokenBinding)
	}
	...
}
```

## Summary

We have taken a quick tour of implementing a `detect` binary using `libcnb`.  We have seen how `libcnb` offers a straightforward `DetectResult` that can be used to signal success or failure of the detect phase.  In addition the `DetectResult` can offer `Provides` and specify `Requires`.  Our tour of the `libcnb` API included a look at `libcnb.DetectContext` through which we have access to environment variables, buildpack metadata and bindings.
