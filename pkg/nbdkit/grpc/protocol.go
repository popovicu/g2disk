package protocol

import (
	"fmt"

	"libguestfs.org/nbdkit"
)

type G2DiskPlugin struct {
	nbdkit.Plugin
	disk []byte
}

type G2DiskConnection struct {
	nbdkit.Connection
	pluginHandle *G2DiskPlugin
}

func (p *G2DiskPlugin) Config(key string, value string) error {
	return nbdkit.PluginError{Errmsg: fmt.Sprintf("unknown parameter %s", key)}
}

func (p *G2DiskPlugin) ConfigComplete() error {
	// return nbdkit.PluginError{Errmsg: "not yet implemented"}
	return nil
}

func (p *G2DiskPlugin) GetReady() error {
	// Allocate the RAM disk. TODO: remove later
	p.disk = make([]byte, 5*1024*1024 /* 5 MB */)
	return nil
}

func (p *G2DiskPlugin) Open(readonly bool) (nbdkit.ConnectionInterface, error) {
	return &G2DiskConnection{
		pluginHandle: p,
	}, nil
}

func (c *G2DiskConnection) GetSize() (uint64, error) {
	return uint64(len(c.pluginHandle.disk)), nil
}

func (c *G2DiskConnection) CanMultiConn() (bool, error) {
	// Clients are not allowed to make multiple connections safely.
	// TODO: reconsider this
	return false, nil
}

func (c *G2DiskConnection) PRead(buf []byte, offset uint64,
	flags uint32) error {
	disk := c.pluginHandle.disk
	copy(buf, disk[offset:int(offset)+len(buf)])
	return nil
}

func (c *G2DiskConnection) CanWrite() (bool, error) {
	return true, nil
}

func (c *G2DiskConnection) PWrite(buf []byte, offset uint64,
	flags uint32) error {
	disk := c.pluginHandle.disk
	copy(disk[offset:int(offset)+len(buf)], buf)
	return nil
}
