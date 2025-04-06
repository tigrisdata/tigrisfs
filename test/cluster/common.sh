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

set -o pipefail
set -x

env|grep NO_PROXY

if [ "$NO_PROXY" != "" ]; then
  . `dirname $0`/s3.sh
else
  . `dirname $0`/proxy.sh
fi

. `dirname $0`/mount.sh

TEST_ARTIFACTS="${TEST_ARTIFACTS:-executions/$(date +"%Y-%m-%d-%H-%M-%S")}"
mkdir -p "$TEST_ARTIFACTS"

_check_nolog() {
  echo "=== S3 bucket setup"
  _s3_setup

  echo "=== Cluster setup"
  _cluster_setup

  echo "=== Test"
  (set -ex; _test)
  EXIT_CODE=$?
  echo "=== Test, exit code = $EXIT_CODE"

  _cleanup

  echo "=== Validate S3"
  (set -ex; _s3_validate)
  S3_VALIDATE_EXIT_CODE=$?
  echo "=== Validate S3, exit code = $S3_VALIDATE_EXIT_CODE"

  _kill_s3proxy

  if [[ $EXIT_CODE == 1 ]]; then
    exit 1
  fi
  if [[ $S3_VALIDATE_EXIT_CODE == 1 ]]; then
    exit 1
  fi
  exit 0
}

_check() {
  _check_nolog 2>&1 | tee -a "$TEST_ARTIFACTS/test_log"
}
