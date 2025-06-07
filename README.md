# g2disk

**g2disk (Go Giant Disk)** is a framework for enabling potentially giant block devices (disks) in userspace for Linux, via [Linux NBD](https://docs.kernel.org/admin-guide/blockdev/nbd.html).

> :warning: This project is not ready for any sort of production use at this point.
> Use with caution for any non-experimental setting and please file requests for any features you'd like to see for production use.
> At the moment, consider this codebase only a proof of concept.

The concept is the following:
1) Build on top of the [nbdkit](https://libguestfs.org/nbdkit.1.html) plugin framework in Go (via `cgo`)
2) Define a gRPC protocol (perhaps other protocols are to follow, e.g. REST) that enables the plugin to proxy the disk requests to a gRPC (or other) service via network to another server which doesn't necessarily need to know about the Linux NBD protocol. The benefit of this is that the target server can be implemented in more modern server frameworks without needing to do anything difficult to wire in with the NBD set up.
3) Implement the aforementioned target server.
4) Build the `g2disk` plugin `.so` file from this repo and start `nbdkit` with it.
5) Connect your Linux `nbd-client` to the `nbdkit` instance from the previous step, and that instance can proxy over to your target gRPC server.

At the moment, there is only a reference implementation for a gRPC-based plugin proxy and the corresponding Go "ramdisk" server. Please feel free to contribute new implementations and file requests for other implementations.

As mentioned above, the current implementation available is just a proof of concept and doesn't even run gRPC over TLS. If there is any interest in using this project where this would actually matter, please file a feature request.

# Docs

This file only contains the instructions for the quickest way to get started. For detailed documentation, head over to [the docs index](/docs/index.md).

# How to use

This intended build tool for this repository is `bazel`.

## Building the `nbdkit` plugin

Simply run:

```
bazel build //pkg/nbdkit/grpc:g2disk
```

> :warning: If you're facing linker issues, consider using the `linkopt` flag like below, for example.
> If you customize the linking with `linkopt`, you may want to use your flag for all the builds as your Bazel flag otherwise gets discarded.

```
bazel build --linkopt=-fuse-ld=gold //pkg/nbdkit/grpc:g2disk
```

The build should produce a file `libg2disk.so`.

## Using the `nbdkit` plugin

`nbdkit` needs to be used in the foreground (`-f`) mode because the Go plugin is used. Something like this should work:

```
sudo nbdkit -f -U /tmp/g2disk.sock bazel-bin/pkg/nbdkit/grpc/g2disk_/libg2disk.so endpoint=localhost:8080
```

That runs `nbdkit` server instance that listens on a Unix socket. Alternatively, the server can listen on a TCP socket if set up that way.

The `.so` file parameter is the newly built plugin that will proxies NBD requests over to the endpoint, defined at the end of the line.

You should now see output like this:

```
2025/06/01 16:57:38 INFO Configuration complete
2025/06/01 16:57:38 INFO Creating a gRPC client
```

## Creating an NBD drive

Next, you use `nbd-client` to wire things in the kernel.

```
sudo nbd-client -unix /tmp/g2disk.sock /dev/nbd1
```

The output should be something like:

```
Warning: the oldstyle protocol is no longer supported.
This method now uses the newstyle protocol with a default export
Negotiation: ..size = 5MB
Connected /dev/nbd1
```

You can use other `/dev/ndb*` devices, of course.

## Using the NBD device

You can then proceed to use your NBD device just like any other block device.

```
sudo mkfs.ext4 /dev/nbd1
```

That will format the device and you're ready to mount it:

```
sudo mount /dev/nbd1 /mnt/disk
```

From that point on, your `/mnt/disk` directory is a regular `ext4` mountpoint.

## Stopping the NBD device

After you're done with the filesystem, unmount the filesystem and stop the device with something like this:

```
sudo nbd-client -d /dev/nbd1
```

You can then also `Ctrl-C` the main `nbdkit` server process.