load("@rules_go//go:def.bzl", "go_path")

go_path(
    name = "gopath",
    deps = [
        "//cmd/server/grpc/ramdisk:ramdisk_server",
        "@nbdkit//plugins/golang/src/libguestfs.org/nbdkit:nbdkit_plugin",
    ],
)
