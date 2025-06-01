load("//:nbdkit.bzl", "nbdkit_repo")

def _nbdkit_impl(_mctx):
    nbdkit_repo(
        name = "nbdkit", # TODO: consider not hardcoding
    )


nbdkit = module_extension(
    implementation = _nbdkit_impl,
)