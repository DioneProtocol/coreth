#!/usr/bin/env bash

set -euo pipefail

# Run OdysseyGo e2e tests from the target version against the current state of coreth.

# e.g.,
# ./scripts/tests.e2e.sh
# AVALANCHE_VERSION=v1.10.x ./scripts/tests.e2e.sh
if ! [[ "$0" =~ scripts/tests.e2e.sh ]]; then
  echo "must be run from repository root"
  exit 255
fi

# Coreth root directory
CORETH_PATH=$( cd "$( dirname "${BASH_SOURCE[0]}" )"; cd .. && pwd )

# Allow configuring the clone path to point to an existing clone
AVALANCHEGO_CLONE_PATH="${AVALANCHEGO_CLONE_PATH:-odysseygo}"

# Load the version
source "$CORETH_PATH"/scripts/versions.sh

# Always return to the coreth path on exit
function cleanup {
  cd "${CORETH_PATH}"
}
trap cleanup EXIT

echo "checking out target OdysseyGo version ${odyssey_version}"
if [[ -d "${AVALANCHEGO_CLONE_PATH}" ]]; then
  echo "updating existing clone"
  cd "${AVALANCHEGO_CLONE_PATH}"
  git fetch
  git checkout -B "${odyssey_version}"
else
  echo "creating new clone"
  git clone -b "${odyssey_version}"\
      --single-branch https://github.com/DioneProtocol/odysseygo.git\
      "${AVALANCHEGO_CLONE_PATH}"
  cd "${AVALANCHEGO_CLONE_PATH}"
fi

echo "updating coreth dependency to point to ${CORETH_PATH}"
go mod edit -replace "github.com/DioneProtocol/coreth=${CORETH_PATH}"
go mod tidy

echo "building odysseygo"
./scripts/build.sh -r

echo "running OdysseyGo e2e tests"
./scripts/tests.e2e.sh ./build/odysseygo
