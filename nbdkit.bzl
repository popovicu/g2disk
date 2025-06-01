def _impl_local(rctx):
    rctx.download_and_extract(
        url = "https://gitlab.com/nbdkit/nbdkit/-/archive/stable-1.38/nbdkit-stable-1.38.tar.gz",
        strip_prefix = "nbdkit-stable-1.38",
        output = ".",
    )

    # TODO: do not hardcode major, minor, micro
    rctx.file(
        "./include/BUILD",
        content = """load("@g2disk//:version.bzl", "expand_version")

package(
    default_visibility = [
        "//visibility:public",
    ],
)

expand_version(
    name = "version_header_expanded",
    major = "1",
    minor = "16",
    micro = "2",
    expanded_file = ":nbdkit-version.h",
    template = ":nbdkit-version.h.in",
)

exports_files(["nbdkit-common.h", "nbdkit-filter.h", "nbdkit-plugin.h"])
""")

    rctx.file(
        "./BUILD",
        content = """cc_library(
    name = "nbdkit",
    hdrs = [
        "//include:nbdkit-common.h",
        "//include:nbdkit-filter.h",
        "//include:nbdkit-plugin.h",
        "//include:nbdkit-version.h",
    ],
    includes = [
        "include",
    ],
    visibility = ["//visibility:public"],
)
""")

    rctx.file(
        "./plugins/golang/src/libguestfs.org/nbdkit/BUILD",
        content = """load("@rules_go//go:def.bzl", "go_library")

go_library(
    name = "nbdkit_plugin",
    srcs = [
        "nbdkit.go",
        "utils.go",
        "wrappers.go",
        "wrappers.h",
    ],
    cdeps = [
        "//:nbdkit",
    ],
    cgo = True,
    importpath = "libguestfs.org/nbdkit",
    visibility = ["//visibility:public"],
)
""")

nbdkit_repo = repository_rule(
    implementation = _impl_local,
    attrs = {},
)