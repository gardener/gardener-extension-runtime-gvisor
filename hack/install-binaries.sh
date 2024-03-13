#!/usr/bin/env sh
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -e

GVISOR_VERSION=$1

# Install runsc (gVisor) and containerd-shim-runsc-v1 (shim for gVisor)
ARCH=$(uname -m)
URL="https://storage.googleapis.com/gvisor/releases/release/${GVISOR_VERSION}/${ARCH}"
wget "${URL}/runsc" "${URL}/runsc.sha512" \
    "${URL}/containerd-shim-runsc-v1" "${URL}/containerd-shim-runsc-v1.sha512"
sha512sum -c runsc.sha512 \
    -c containerd-shim-runsc-v1.sha512
rm -f -- *.sha512
chmod a+rx runsc containerd-shim-runsc-v1
mv runsc containerd-shim-runsc-v1 /usr/local/bin
