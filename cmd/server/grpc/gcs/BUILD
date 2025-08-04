load("@rules_go//go:def.bzl", "go_binary")

# To test:
#   go install github.com/fsouza/fake-gcs-server@latest
#   fake-gcs-server -scheme http -host 127.0.0.1 -port 4443 -public-host 127.0.0.1:4443 -external-url http://127.0.0.1:4443 -filesystem-root /tmp/gcs
#   bazel run --linkopt=-fuse-ld=gold //cmd/server/grpc/gcs:gcs_server -- --gcs-endpoint=http://127.0.0.1:4443/storage/v1/ --bucket=g2disk --port=9876 --auto-create-bucket
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
        "//remote/gcs",
        "//remote/google_api:iterator",
        "//remote/google_api:option",
        "//remote/grpc",
        "//remote/grpc:grpc_codes",
        "//remote/grpc:grpc_status",
    ],
)
