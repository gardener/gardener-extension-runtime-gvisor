#!/bin/bash
#
# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

RUNSC_VERSION=$1
CONTAINERD_RUNSC_SHIM_VERSION=$2

# Install runsc (gvisor)
URL=https://storage.googleapis.com/gvisor/releases/release/$RUNSC_VERSION
wget ${URL}/runsc
wget ${URL}/runsc.sha512
sha512sum -c runsc.sha512
rm -f runsc.sha512
mv runsc /usr/local/bin
chmod 0755 /usr/local/bin/runsc

# Install runsc containerd shim
URL=https://github.com/google/gvisor-containerd-shim/releases/download/$CONTAINERD_RUNSC_SHIM_VERSION
wget ${URL}/containerd-shim-runsc-v1.linux-amd64
mv containerd-shim-runsc-v1.linux-amd64 /usr/local/bin
chmod 0755 /usr/local/bin/containerd-shim-runsc-v1.linux-amd64