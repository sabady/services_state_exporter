#!/bin/bash

# --no-cache --rm --force-rm
#
set -e

MANIFEST_NAME="go-swarm-exporter"
if [ -z $BUILD_NUMBER ]; then 
  VERSION=$(cat version.txt)
else 
  VERSION=$(cat version.txt)_${BUILD_NUMBER}
fi

podman build \
  --jobs=1 \
  --format=docker \
  --layers=false \
  --platform="linux/arm64,linux/amd64" \
  --pull \
  --progress=plain \
  --compress \
  -f Dockerfile \
  --manifest="${MANIFEST_NAME}:${VERSION}" \
  .

podman manifest push \
  --all ${MANIFEST_NAME}:${VERSION} docker://${MANIFEST_NAME}:${VERSION}

trivy image --exit-code 0 --severity LOW,MEDIUM,HIGH,CRITICAL ${MANIFEST_NAME}:${VERSION}

# vim:set ts=2 sw=2 sts=2 et :
