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
  ActionsLexer.g4 ActionsParser.g4

sed -i 's/interface{}/any/g' *.go
sed -i 's/ActionsParserVisitor/ActionsVisitor/g' *.go
sed -i 's/ActionsParserListener/ActionsListener/g' *.go
sed -i 's/ActionsParserParserStaticData/ActionsParserStaticData/g' *.go
sed -i 's/ActionsLexerLexerStaticData/ActionsLexerStaticData/g' *.go

mv actionsparser_visitor.go       actions_visitor.go
mv actionsparser_base_visitor.go  actions_base_visitor.go
mv actionsparser_listener.go      actions_listener.go
mv actionsparser_base_listener.go actions_base_listener.go

go fmt .
popd
