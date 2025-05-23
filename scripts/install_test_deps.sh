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

export GO111MODULE=on

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)/bin" latest

if [ "$(uname -s)" = "Darwin" ]; then
  if command -v brew > /dev/null 2>&1; then
    brew install shellcheck
  fi
else
  sudo apt-get install -y shellcheck \
    s3cmd \
    util-linux \
    fuse3
fi
