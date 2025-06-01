load("@rules_go//go:def.bzl", "go_binary", "go_library")

PLUGIN_LIB = "@nbdkit//plugins/golang/src/libguestfs.org/nbdkit:nbdkit_plugin"

go_library(
    name = "protocol",
    srcs = [
        "protocol.go",
    ],
    deps = [
        PLUGIN_LIB,
        "//remote/grpc",
        "//remote/grpc:grpc_credentials",
        "//remote/grpc:grpc_credentials_insecure",
        "//proto:g2disk_service_go_proto",
    ],
    cgo = True,
    importpath = "github.com/popovicu/g2disk/pkg/nbdkit/grpc/protocol",
)

go_binary(
    name = "g2disk",
    deps = [
        ":protocol",
        PLUGIN_LIB,
    ],
    srcs = [
        "g2disk.go",
    ],
    cgo = True,
    linkmode = "c-shared",
    visibility = [
        "//:__pkg__",
    ]
)