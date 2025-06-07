# Building

This project uses Bazel and the recommended way to leverage this is via [Bazelisk](https://github.com/bazelbuild/bazelisk). It requires no installation, just a single binary download from GitHub.

For a high level overview of Bazel, please check out [this text](https://popovicu.com/posts/build-all-software-in-one-command-with-bazel/).

## Requirements

Bazelisk binary (ideally aliased as `bazel`) should be enough to run the Bazel build flows. In addition, a working C compiler is needed which can target C standard library (basically any C compiler). Other than that, the Bazel build flow will dynamically fetch the Go toolchain, protobuf compiler, etc.

## Handling binaries

Unless you know exactly what you're doing, it's highly recommended that you build the `.so` file on the same environment where you want to run `nbdkit`. At the moment, the `.so` file dynamically links to the C standard library from the build machine, so any inconsistency between the build library and the target can result in problems. Work is in progress to turn this into a fully static build.

The statically linked (especially pure Go) servers are much more portable and can be built anywhere and distributed without issues.

## Server containers

Ultimately, the project should have targets for easily building server container images that are trivial to build and distributed, as described [here](https://popovicu.com/posts/containers-bazel-one-command/).