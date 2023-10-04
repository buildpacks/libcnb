# `libcnb` tutorial

[https://pkg.go.dev/github.com/buildpacks/libcnb?tab=doc](https://pkg.go.dev/github.com/buildpacks/libcnb?tab=doc) provides the reference documentation for `libcnb`.  In this tutorial we walk through an example implementation using data structures and functions described in the reference documentation.

In general, buildpacks are responsible for

* installing runtimes eg: a `python` interpreter,
* installing dependency managers eg: Python's `pip` dependency manager,
* resolving application dependencies, and
* building/packaging the application code.

Commonly each of these responsibilities is factored out into a separate buildpacks.

In this tutorial we take a simplified look at a generic buildpack and implement both the `detect` and `build` binaries for an example buildpack.

## Scenario

Our buildpack will have the following properties:

* if the source project contains `example.txt` or `BP_EXAMPLE_FILE=some-file.txt` then the buildpack `detect` process will succeed.
* if `detect` passes, then our `build` binary will contribute a single layer to the output image.  The layer will contain the contents of a gzipped archive fetched from an Internet location.

We first [`detect`](detect.md) whether this buildpack should contribute to the build.  If detection passes we then run [`build`](build.md) to extract our example gzipped archive.

The main application of our buildpack will combine our implementation of `Detector` and `Builder` to provide the `detect` and `build` binaries:

```go
func main() {
	detector := Detector{Logger: log.New(os.Stdout)}
	builder := Builder{Logger: log.New(os.Stdout)}
	libcnb.Main(detector, builder)
}
```

## Learning Goals

At the end of this tutorial you will have

* examined the separation of `detect` responsibilities from `build` responsibilities,
* introspected a source application to satisfy `detect`,
* passed metadata between the `detect` and `build` phases using the build plan,
* contributed a layer to an output image that contains the contents of an archive,
* set a default environment variable `EXAMPLE` on the output image, and
* defined an exec.d startup script on the output image.
