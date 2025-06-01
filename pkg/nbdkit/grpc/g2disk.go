package main

import (
	"C"
	"unsafe"

	"libguestfs.org/nbdkit"

	"github.com/popovicu/g2disk/pkg/nbdkit/grpc/protocol"
)

//export plugin_init
func plugin_init() unsafe.Pointer {
	// If your plugin needs to do any initialization, you can
	// either put it here or implement a Load() method.
	// ...

	return nbdkit.PluginInitialize("g2disk", &protocol.G2DiskPlugin{})
}

// This is never called, but must exist.
func main() {}
