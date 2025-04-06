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

set -e

VERSION=1.24.0
ARCH=$(dpkg --print-architecture)
FN="go${VERSION}.linux-${ARCH}.tar.gz"

case "$ARCH" in
"amd64")
  SHA256="dea9ca38a0b852a74e81c26134671af7c0fbe65d81b0dc1c5bfe22cf7d4c8858"
  ;;
"arm64")
  SHA256="c3fa6d16ffa261091a5617145553c71d21435ce547e44cc6dfb7470865527cc7"
  ;;
*)
  echo "No supported architecture."
  exit 1
  ;;
esac

wget "https://go.dev/dl/$FN"
echo "$SHA256  $FN" | shasum -a 256 -c

mkdir -p /usr/local
tar -C /usr/local -xzf "$FN"
rm "$FN"

export PATH=$PATH:/usr/local/go/bin

