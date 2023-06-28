# Building

A buildpack is suitably packaged `detect` binary, a `build` binary and a `buildpack.toml`.  Here we look at the implementation logic of the build binary for our buildpack.  This starts by declaring a `struct` that we call `Builder`.  The purpose of our struct is to satisfy `libcnb`s [`Builder`](https://pkg.go.dev/github.com/buildpacks/libcnb?utm_source=gopls#Builder) interface providing a receiver with signature `func Build(context BuildContext) (BuildResult, error)`.  An implementation of `Builder` is passed to `libcnb.Main` as the main entry point to the `build` binary.

In our `Detector` we use a [`poet`](https://pkg.go.dev/github.com/buildpacks/libcnb/poet) logger instance.  The logger instance will be used to inform the user of the progress of our `detect` phase.

```go
type Builder struct {
	Logger poet.Logger
}
```

## Implementing Build

The most simple implementation of `func (context BuildContext) (BuildResult, error)` returns a `BuildResult` containing an ordered list `LayerContributor`s which will each contribute layers to the image.  

```go
func (d Builder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	return libcnb.NewBuildResult(), nil
}
```

The trivial passing implementation of `Build` contributes no layers to the result.

This trivial case ignores the `Plans` field of the `DetectResult` struct.  In order to understand the `Plans` field we look at the concept of provides and requires.

### Builders and Contributors

It is recommended that a single `Builder` provides each layer using a `Contributor`.  In this example we are contributing a single layer containing our extracted archive.  Therefore, our `Builder` will orchestrate a single `Contributor`.

### Accessing `BuildPlan` Metadata

The `detect` phases passes metadata to the `build` phase in the [`BuildpackPlan`](https://pkg.go.dev/github.com/buildpacks/libcnb#BuildpackPlan) (note: not to be confused with the more general [`BuildPlan`](https://pkg.go.dev/github.com/buildpacks/libcnb#BuildPlan)).  Multiple buildpack `Detect` executions can require that the `Build` of our example buildpack executes.  As there may be requirements from multiple `detect` binaries, we must merge all the entries in the buildplan that correspond to our "example" buildpack.  To resolve the build plan we choose to use a [utility function from `libpak`](https://github.com/paketo-buildpacks/libpak/blob/f24422191fc6a2a02178337d96dda1210faaae9f/buildpack_plan.go#L69).

```go
func (b Builder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
    // Resolve the "version" that the detect phase adds to the BuildPlan metadata
	version, err := resolveVersion(context.Plan)
	if err != nil {
		return libcnb.NewBuildResult(), err
	}
    // If "version" was not supplied, use a default
	if version == "" {
		version = "0.0.1"
	}

    return libcnb.NewBuildResult(), nil
}

// resolveVersion first merges all libcnb.BuildpackPlan entries releated to the
// "example" buildpack.  It then returrns the "version" information extracted
// from the metadata.
func resolveVersion(plan libcnb.BuildpackPlan) (string, error) {
	resolver := libpak.PlanEntryResolver{Plan: plan}
	resolvedPlan, resolved, err := resolver.Resolve("example")
	if err != nil {
		return "", fmt.Errorf("unable to resolve version from metadata. Error %w", err)
	}
	if !resolved {
		return "", nil
	}
	if version, ok := resolvedPlan.Metadata["version"].(string); ok {
		return version, nil
	}
    return "", nil
}
```

### Parsing the Buildpack Metadata

We download the files to extract into a layer from an archive on the Internet.  To make this easier to maintain, we embed the download URL as a string template in the `builpack.toml` metadata.  Our download URL is of the form `"https://source.fake/releases/v{{ .version }}/example_{{ .version }}_linux_amd64.tar.gz"`.  We accept the version parameter as detected in the `detect` phase, and replace the version into our download URL.

```go
func (b Builder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)
	version, err := resolveVersion(context.Plan)
	if err != nil {
		return libcnb.NewBuildResult(), err
	}
	if version == "" {
		version = "0.0.1"
	}

    // Use the standard Go text/template to compute the download URL from the
    // version resolved from the buildpack plan.
	urlTemplateMetadata, found := context.Buildpack.Metadata["url"]
	if !found {
		return libcnb.NewBuildResult(), fmt.Errorf("no url template in buildpack.toml metadata")
	}
	urlTemplate, ok := urlTemplateMetadata.(string)
	if !ok {
		return libcnb.NewBuildResult(), fmt.Errorf("unable to parse buildpack.toml metadata")
	}
	archiveURL, err := template.New("ArchiveVersion").Parse(urlTemplate)
	if err != nil {
		return libcnb.NewBuildResult(), fmt.Errorf("unable buildpack.toml url metadata is not a valid template")
	}
	var url bytes.Buffer
	archiveURL.Execute(&url, struct{ version string }{version: version})

	return libcnb.NewBuildResult(), nil
}
```

Now that we know our download URL, we are in a position to contribute a layer containing the extracted files.

### Contributing a Layer

To contribute a layer we implement the [`LayerContributor`](https://pkg.go.dev/github.com/buildpacks/libcnb#LayerContributor) interface.  Our `Contributor` will download an archive from the provided URL.  Again, we borrow a utility function from `libpak` to extract the archive to an output layer.

```go
type Contributor struct {
	Logger  log.Logger
	Archive string
}

func (Contributor) Name() string {
	return "example"
}

func (c Contributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
    // the values for caching layers are documented at
    // https://buildpacks.io/docs/buildpack-author-guide/caching-strategies/
    // Here we make the layer available for caching, as a build layer and as a
    // launch layer on the output image
	layer.Launch = true
	layer.Cache = true
	layer.Build = true

	if err := os.MkdirAll(layer.Exec.Path, 0755); err != nil {
		return libcnb.Layer{}, err
	}

	// Fetch the archive from the Internet
	resp, err := http.Get(c.Archive)
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to download archive\n%w", err)
	}
	defer resp.Body.Close()

    // Extract the archive into the output layer
	if err := crush.Extract(resp.Body, layer.Path, 0); err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to expand archive\n%w", err)
	}

	return layer, nil
}
```

We add our contributor to the `libcnb.BuildContext` layers.  `libcnb` will ensure each `LayerContributor` is executed in the order that they are appended to the `BuildResult.Layers` list.

```go
func (b Builder) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	b.Logger.Title(context.Buildpack)

    ...

    contributor := Contributor{
		Logger:  b.Logger,
		Archive: url.String(),
	}
	result := libcnb.NewBuildResult()
	result.Layers = append(result.Layers, contributor)
	return result, nil
}
```

### Contributing a Startup Script and Environmental Variable

Finally, we assume that our run image contains a `bash` shell.  An [`exec.d` style](https://buildpacks.io/docs/reference/spec/migration/buildpack-api-0.4-0.5/#execd) executable is contributed as to run on container startup.

An environment variable `EXAMPLE` is contributed to the image with the default value of `1.2.3`.  It is left as an exercise to the reader to pass the `version` information from `Build` through the instance of `type Contributor struct` to the `EXAMPLE` environmental variable.

```go
const ExampleExecD string = `
#!/bin/bash

echo Executed ${EXAMPLE} on Startup
`

func (c Contributor) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	layer.Launch = true
	layer.Cache = true
	layer.Build = true

	if err := os.MkdirAll(layer.Exec.Path, 0755); err != nil {
		return libcnb.Layer{}, err
	}

	// fetch and extract tar.gz
	resp, err := http.Get(c.Archive)
	if err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to download archive\n%w", err)
	}
	defer resp.Body.Close()
	if err := crush.Extract(resp.Body, layer.Path, 0); err != nil {
		return libcnb.Layer{}, fmt.Errorf("unable to expand archive\n%w", err)
	}

	// We are using the exec.d interface here per - https://github.com/buildpacks/spec/blob/main/buildpack.md#execd
	// #nosec G306
	if err := os.WriteFile(layer.Exec.FilePath("example"), []byte(ExampleExecD), 0755); err != nil {
		return libcnb.Layer{}, err
	}

	// Set environment variable
	layer.LaunchEnvironment.Default("EXAMPLE", "1.2.3")

	return layer, nil
}
```

## Summary

We have taken a quick tour of implementing a `build` binary using `libcnb`.  We have seen how `libcnb` offers a straightforward `BuildResult` that can be used to return a number of `LayerContributor`s.  In turn, each `LayerContributor` can contribute files to the output image, scripts executed on startup and environmental variable on the output image.  Our tour of the `libcnb` API included a look at `libcnb.BuildContext` through which we have access to the `BuildPlan` computed after the `detect` phase and buildpack metadata.
