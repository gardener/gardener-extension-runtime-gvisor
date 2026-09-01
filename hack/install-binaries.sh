#!/usr/bin/env sh
#
# SPDX-FileCopyrightText: Contributors to the Gardener project
#
# SPDX-License-Identifier: Apache-2.0

set -e

GVISOR_VERSION=$1

# Install runsc (gVisor), containerd-shim-runsc-v1 (shim for gVisor), and gvisor binaries from tar file
ARCH=$(uname -m)
URL="https://storage.googleapis.com/gvisor/releases/release/${GVISOR_VERSION}/${ARCH}"
wget "${URL}/gvisor.tar.bz2" "${URL}/gvisor.tar.bz2.sha512"
sha512sum -c gvisor.tar.bz2.sha512
tar -vxjf gvisor.tar.bz2 -C /usr/local/bin
