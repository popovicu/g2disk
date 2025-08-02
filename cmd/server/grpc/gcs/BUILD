load("@rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "gcs_server",
    srcs = [
        "gcs_server.go",
    ],
    pure = "on",
    visibility = [
        "//:__pkg__",
    ],
    deps = [
        "//proto:g2disk_service_go_proto",
        "//remote/grpc",
    ],
)
