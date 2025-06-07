# Internals

Below are the explanations for the main components of the codebase.

## MODULE.bazel

`/MODULE.bazel` is the first file to go through in order to understand the internals.

First, a couple of Bazel dependencies are set up: mainly importing the Go toolchain and Bazel rules, as well as the [protobuf](https://protobuf.dev/) Bazel dependencies for the gRPC implementation.

More importantly, a custom `nbdkit` Bazel module extension is used to import the `nbdkit` project into the build flow here.

## nbdkit module extension

`/extension.bzl` simply adds a new module/repository called `nbdkit` (currently hardcoded) into the build flow. The inner mechanics are in `nbdkit.bzl`.

`nbdkit.bzl` has the `nbdkit_repo` repo rule which downloads a stable version of the `nbdkit` source code and inserts several `BUILD` files. Theoretically, Gazelle should be used for this to handle it automatically, but the project is organized slightly specially which prevents Gazelle from working correctly here, and thus the `BUILD` files are manually injected.

The first thing that happens here is the header file `nbdkit-version.h` gets its template expanded. The `nbdkit` source code has a template which leaves a placeholder for setting the right API version and we use a trivial rule from `/version.bzl` (in this codebase) to expand it. That code should be self-explanatory.

Next, a `cc_library` Bazel target is injected into the `nbdkit` codebase to capture the library needed to write C-based `nbdkit` plugins. This library consists of only the C header files.

That library is then used in the relevant `go_library` via `cgo` and thus we get a Go Bazel library for writing `nbdkit` plugins.

## Go plugins

From this point, it's trivial to build Go-based plugins as the `nbdkit` Go API is now available. The library targets in the `/pkg/nbdkit/...` locations are self-explanatory. `go_binary` with the right `linkmode` value produces the `.so` file that can be loaded as an `nbdkit` plugin.