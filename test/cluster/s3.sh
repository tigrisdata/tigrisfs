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

echo "=== Use S3 ==="

#unset ENDPOINT

env|grep ENDPOINT

TEST_ARTIFACTS="${TEST_ARTIFACTS:-executions/$(date +"%Y-%m-%d-%H-%M-%S")}"
mkdir -p "$TEST_ARTIFACTS"
export S3CMD_ARGS="--access_key=foo --secret_key=bar"

_kill_s3proxy() {
    :
}

_s3cmd() {
  s3cmd \
  $S3CMD_ARGS \
  --signature-v2 \
  --no-ssl \
  --host-bucket="" \
  --host="$ENDPOINT" \
  "$@"
}

_aws_cli() {
  aws s3 --endpoint-url="$ENDPOINT" "$@"
}

export BUCKET_NAME="test.$RANDOM"

_aws_cli mb s3://$BUCKET_NAME

