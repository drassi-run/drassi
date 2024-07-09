#!/bin/bash

set -eux

DIR=$(dirname "$(realpath "${BASH_SOURCE[0]}")")
IMAGE=drassi.run/antlr4-tool:4.13

pushd "$DIR"
docker build -t "$IMAGE" -f "./Dockerfile" .
docker run --rm -v "$DIR:/work" -u "$(id -u):$(id -g)" "$IMAGE" \
  -Dlanguage=Go \
  -package grammar \
  -visitor \
  GHALexer.g4

go fmt .
popd
