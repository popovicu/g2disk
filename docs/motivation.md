# Motivation

Having the ability to build Linux block devices in userspace should be an obvious win. The userspace has access to numerous extremely powerful storage APIs that can distribute, secure and otherwise transport the data across different distributed systems, and nowadays it is very easy to do so. Additionally, cloud providers like Google, AWS, etc. make cloud storage cheap and reliable.

NBD, however, is not the only way to acheive this on Linux. There are different components of the kernel that can achieve similar goals, but NBD seems to be the simplest to get started with.

Also, NBD is generally a simple protocol, and one of the options was to not rely on `nbdkit` at all and simply implement the protocol management in a custom binary. However, even though the protocol is simple, at least in comparison to other widely used protocols out there, the original author decided this was outside the scope of the project and simply relied on `nbdkit` which itself is developed by Red Hat and gets regular code contributions. In short, Red Hat probably does a much better job at maintaing the protocol handling than the original author ever could on his own. :)