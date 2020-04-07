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


source $(dirname "${0}")/helper.sh
source $(dirname "${0}")/local-imagevector-overwrite.sh

while getopts "l:d:e:i:" opt
do
   case "$opt" in
      l ) LD_FLAGS="$OPTARG" ;;
      d ) DIRECTORY="$OPTARG" ;;
      i ) IGNORE_OPERATION_ANNOTATION="$OPTARG" ;;
      ? ) gvisorParamterUsage ;; # Print gvisorParamterUsage in case parameter is non-existent
   esac
done

# Print gvisorParamterUsage in case parameters are empty
if [ -z "$LD_FLAGS" ] || [ -z "$DIRECTORY" ]
then
   echo "Some or all of the parameters are empty";
   gvisorParamterUsage
fi

# Begin script in case all parameters are correct
kubeconfig="$(mktemp_kubeconfig)"
trap cleanup_kubeconfig EXIT

imagevector_overwrite="$(mktemp_imagevector_overwrite)"
trap cleanup_imagevector_overwrite EXIT

KUBECONFIG="${KUBECONFIG:-$kubeconfig}" \
IMAGEVECTOR_OVERWRITE="$imagevector_overwrite" \
GO111MODULE=on \
    go run \
      -mod=vendor \
      -ldflags "$LD_FLAGS" \
      "$DIRECTORY" \
      --ignore-operation-annotation="$IGNORE_OPERATION_ANNOTATION" \
      --leader-election=false