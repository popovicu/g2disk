load("@rules_go//go:def.bzl", "go_binary", "go_library")

PLUGIN_LIB = "@nbdkit//plugins/golang/src/libguestfs.org/nbdkit:nbdkit_plugin"

go_library(
    name = "protocol",
    srcs = [
        "protocol.go",
    ],
    cgo = True,
    importpath = "github.com/popovicu/g2disk/pkg/nbdkit/grpc/protocol",
    deps = [
        PLUGIN_LIB,
        "//proto:g2disk_service_go_proto",
        "//remote/grpc",
        "//remote/grpc:grpc_credentials",
        "//remote/grpc:grpc_credentials_insecure",
    ],
)

go_binary(
    name = "g2disk",
    srcs = [
        "g2disk.go",
    ],
    cgo = True,
    linkmode = "c-shared",
    visibility = [
        "//:__pkg__",
    ],
    deps = [
        ":protocol",
        PLUGIN_LIB,
    ],
)
