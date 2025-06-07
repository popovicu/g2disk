# API

The goal of this repository is to provide ready-to-use binaries for different use cases, and ideally it would be a monorepo of different flavors of the `nbdkit`, as well as different servers, however, the project also exposes some intermediate Bazel-related internals to build something completely different.

As mentioned in the internals document, the `nbdkit` codebase itself is injected into the build flow, and Bazel build definitions are inserted on the fly. This exposes both the C and Go APIs to the rest of the Bazel build flow.

If you want to build your own concept of `nbdkit` plugins outside of this repository, you can simply define a dependency on this codebase in Bazel and use the exposed targets directly.