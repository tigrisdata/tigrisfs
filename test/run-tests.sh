#!/bin/bash
# Copyright 2025 Tigris Data, Inc.
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

#set -o xtrace
set -o errexit

T=
if [ $# == 1 ]; then
    T="-check.f $1"
fi

if [ "$NO_PROXY" == "" ]; then
  . "$(dirname "$0")/run-proxy.sh"
fi

if [ "$TIMEOUT" != "" ]; then
  TIMEOUT="-test.timeout $TIMEOUT"
fi

# run test in `go test` local mode so streaming output works
cd core
CGO_ENABLED=1 go test -race -v $TIMEOUT -check.vv $T
exit $?
