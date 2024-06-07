#!/bin/bash -e
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

gvisorParameterUsage()
{
   echo ""
   echo "Usage: $0 -l LD_FLAGS -d DIRECTORY -i IGNORE_OPERATION_ANNOTATION -r REPO_ROOT"
   echo -e "\t-l ldflags for the Go compilation"
   echo -e "\t-d Directory to the go main() function"
   echo -e "\t-i Wether to ignore the operation annotation on ContainerRuntime resources"
   echo -e "\t-r Filepath to the root of the git repository"
   exit 1
}

while getopts "l:d:e:i:r:" opt
do
   case "$opt" in
      l ) LD_FLAGS="$OPTARG" ;;
      d ) DIRECTORY="$OPTARG" ;;
      i ) IGNORE_OPERATION_ANNOTATION="$OPTARG" ;;
      r ) REPO_ROOT="$OPTARG" ;;
      ? ) gvisorParameterUsage ;; # Print gvisorParameterUsage in case parameter is non-existent
   esac
done

# Print gvisorParameterUsage in case parameters are empty
if [ -z "$LD_FLAGS" ] || [ -z "$DIRECTORY" ] ||  [ -z "$IGNORE_OPERATION_ANNOTATION" ] ||  [ -z "$REPO_ROOT" ]
then
   echo "Some or all of the parameters are empty";
   gvisorParameterUsage
fi

echo "Using LD_FLAGS: $LD_FLAGS"
echo "Ignoring operation annotation: $IGNORE_OPERATION_ANNOTATION"

# contains common helper functions (needed: mktemp_kubeconfig() & cleanup_kubeconfig)
source "${GARDENER_HACK_DIR}/local-development/common/helpers
source "${GARDENER_HACK_DIR}/local-development/common/local-imagevector-overwrite

# Begin script in case all parameters are correct
kubeconfig="$(mktemp_kubeconfig)"
trap cleanup_kubeconfig EXIT

file_imagevector_overwrite="$(mktemp_imagevector_overwrite github.com/gardener/gardener-extension-runtime-gvisor "$REPO_ROOT" "$REPO_ROOT"/charts)"
local_image_vector=$(cat "$file_imagevector_overwrite")
echo "Local image vector override: $local_image_vector"

if [ ! -f "$file_imagevector_overwrite" ]; then
    echo "failed to generate local image vector override: $file_imagevector_overwrite"
else
  trap cleanup_imagevector_overwrite EXIT

  KUBECONFIG="${KUBECONFIG:-$kubeconfig}" \
  IMAGEVECTOR_OVERWRITE="$file_imagevector_overwrite" \
  GO111MODULE=on \
      go run \
        -ldflags "$LD_FLAGS" \
        "$DIRECTORY" \
        --ignore-operation-annotation="$IGNORE_OPERATION_ANNOTATION" \
        --leader-election=false \
        --gardener-version="v1.39.0"
fi
