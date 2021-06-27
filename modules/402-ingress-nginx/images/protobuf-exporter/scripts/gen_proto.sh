#!/bin/bash

# Copyright 2021 Flant CJSC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

if ! [[ "$0" =~ "scripts/gen_proto.sh" ]]; then
  echo "must be run from repository root"
  exit 255
fi

if ! [[ $(protoc --version) =~ "3.12.0" ]]; then
  echo "could not find protoc 3.12.0, is it installed + in PATH?"
  exit 255
fi

echo "Installing gogo/protobuf..."
GOGOPROTO_ROOT="$GOPATH/src/github.com/gogo/protobuf"
GO111MODULE="off" go get -v github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto,protoc-gen-gofast}
GO111MODULE="off" go get -v golang.org/x/tools/cmd/goimports

ln -s $GOPATH/bin/protoc-gen-gofast /usr/local/bin/protoc-gen-gogofast || true

echo "Generating message"
protoc --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf:. --gogofast_out=. ./pkg/proto/*.proto
