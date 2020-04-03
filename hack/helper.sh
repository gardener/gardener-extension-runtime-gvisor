#!/bin/bash -e
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

__tmp_kubeconfig=""
mktemp_kubeconfig() {
    if [[ "$__tmp_kubeconfig" != "" ]]; then
        echo "$__tmp_kubeconfig"
        return
    fi
    __tmp_kubeconfig="$(mktemp)"
    kubectl config view --raw > "$__tmp_kubeconfig"
    echo "$__tmp_kubeconfig"
}

cleanup_kubeconfig() {
    if [[ "$__tmp_kubeconfig" != "" ]]; then
        rm -f "$__tmp_kubeconfig"
        __tmp_kubeconfig=""
    fi
}

gvisorParamterUsage()
{
   echo ""
   echo "Usage: $0 -l LD_FLAGS -d DIRECTORY"
   echo -e "\t-l ldflags for the Go compilation"
   echo -e "\t-d Directory to the go main() function"
   exit 1 # Exit script after printing help
}