package main

import (
	"C"
	"log/slog"
	"unsafe"

	"libguestfs.org/nbdkit"

	"github.com/popovicu/g2disk/pkg/nbdkit/grpc/protocol"
)

//export plugin_init
func plugin_init() unsafe.Pointer {
	logger := slog.Default()
	return nbdkit.PluginInitialize("g2disk", protocol.NewUninitiatedPlugin(logger))
}

// This is never called, but must exist.
func main() {}
