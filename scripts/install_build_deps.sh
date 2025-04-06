#!/usr/bin/env bash
# Copyright 2022-2025 Tigris Data, Inc.
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

set -ex

# Settings
PROTO_VERSION=25.6
PROTO_RELEASES="https://github.com/protocolbuffers/protobuf/releases"

### Prereqs checks ###
# Check if architecture and OS is supported
# and set environment specifics
ARCH=$(uname -m)
OS=$(uname -s)

case "${OS}-${ARCH}" in
"Darwin-arm64")
  BINARIES="brew curl go"
  ;;
"Darwin-x86_64")
  BINARIES="brew curl go"
  ;;
"Linux-aarch64")
  BINARIES="apt-get curl go"
  ;;
"Linux-arm64")
  BINARIES="apt-get curl go"
  ;;
"Linux-x86_64")
  BINARIES="apt-get curl go"
  ;;
*)
  echo "Unsupported architecture ${ARCH} or operating system ${OS}."
  exit 1
  ;;
esac

# Check if required binaries are available in PATH
for bin in ${BINARIES}; do
  binpath=$(command -v "${bin}")
  if [ -z "${binpath}" ] || ! test -x "${binpath}"; then
    echo "Please ensure that $bin binary is installed and in PATH."
    exit 1
  fi
done

# Install protobuf compiler
case "${OS}" in
"Darwin")
  brew install protobuf
  ;;
"Linux")
  case "${ARCH}" in
  "x86_64")
    PROTO_PKG=protoc-$PROTO_VERSION-linux-x86_64.zip
    ;;
  "aarch64")
    PROTO_PKG=protoc-$PROTO_VERSION-linux-aarch_64.zip
    ;;
  *)
    echo "No supported proto compiler for ${ARCH} or operating system ${OS}."
    exit 1
    ;;
  esac
  ;;
*)
  echo "No supported proto compiler for ${ARCH} or operating system ${OS}."
  exit 1
  ;;
esac

if [ -n "$PROTO_PKG" ]; then
  DOWNLOAD_URL="$PROTO_RELEASES/download/v$PROTO_VERSION/$PROTO_PKG"
  echo "Fetching protobuf release ${DOWNLOAD_URL}"
  curl -LO "$DOWNLOAD_URL"
  sudo unzip "$PROTO_PKG" -d "/usr/local/"
  sudo chmod +x "/usr/local/bin/protoc"
  sudo chmod -R 755 "/usr/local/include/"
  rm -f "$PROTO_PKG"
fi

# Install protobuf
export GO111MODULE=on
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1
