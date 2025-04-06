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

PROXY_PORT=${PROXY_PORT:-8080}

cat <<EOF > s3proxy.properties
s3proxy.endpoint=http://localhost:$PROXY_PORT
s3proxy.authorization=aws-v2
s3proxy.identity=foo
s3proxy.credential=bar
jclouds.provider=transient
jclouds.identity=foo
jclouds.credential=bar
jclouds.regions=us-west-2
EOF

p=`dirname $0`

PROXY_BIN="java -Xmx8g --add-opens java.base/java.lang=ALL-UNNAMED -DLOG_LEVEL=trace -Djclouds.wire=debug -jar s3proxy.jar --properties $p/s3proxy.properties"
export S3CMD_ARGS="--access_key=foo --secret_key=bar"
export AWS_ACCESS_KEY_ID=foo
export AWS_SECRET_ACCESS_KEY=bar
export ENDPOINT=http://localhost:$PROXY_PORT

TEST_ARTIFACTS="${TEST_ARTIFACTS:-executions/$(date +"%Y-%m-%d-%H-%M-%S")}"
mkdir -p "$TEST_ARTIFACTS"

echo "=== Start s3proxy on $ENDPOINT"
$PROXY_BIN > "$TEST_ARTIFACTS/s3proxy_log" &
PROXY_PID=$!
sleep 15

_kill_s3proxy() {
  echo "=== Kill s3proxy"
  kill -9 $PROXY_PID
}

echo "=== Create s3://test"

_s3cmd() {
  s3cmd \
  $S3CMD_ARGS \
  --signature-v2 \
  --no-ssl \
  --host-bucket="" \
  --host="http://localhost:$PROXY_PORT" \
  "$@"
}

_aws_cli() {
  aws s3 --endpoint-url="$ENDPOINT" "$@"
}

_s3cmd mb s3://test

export BUCKET_NAME="test"
