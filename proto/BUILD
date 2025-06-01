load("@rules_go//proto:def.bzl", "go_grpc_library")
load("@rules_proto//proto:defs.bzl", "proto_library")

proto_library(
    name = "g2disk_service_proto",
    srcs = ["g2disk_service.proto"],
)

go_grpc_library(
    name = "g2disk_service_go_proto",
    proto = ":g2disk_service_proto",
    importpath = "github.com/popovicu/g2disk/proto",
    visibility = [
        "//pkg/protocol:__subpackages__",
    ]
)