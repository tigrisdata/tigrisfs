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

set -x

cd $(dirname $0)/..

export CLOUD=${CLOUD:-s3}
PROXY_BIN=$PROXY_BIN
PROXY_PID=$PROXY_PID
PROXY_FS=$PROXY_FS
PROXY_PORT=${PROXY_PORT:-8080}
TIMEOUT=${TIMEOUT:-10m}

trap 'kill -9 $PROXY_PID' EXIT

if [ $CLOUD == "s3" ]; then
    sed 's/$PORT/'$PROXY_PORT'/' < test/s3proxy.properties > test/s3proxy_test.properties
    if [ "$PROXY_FS" != "" ]; then
        mkdir -p /tmp/s3proxy
        echo jclouds.provider=filesystem >>test/s3proxy_test.properties
        echo jclouds.filesystem.basedir=/tmp/s3proxy >>test/s3proxy_test.properties
    fi
    PROXY_BIN="java -Xmx8g --add-opens java.base/java.lang=ALL-UNNAMED -jar s3proxy.jar --properties test/s3proxy_test.properties"
    export AWS_ACCESS_KEY_ID=foo
    export AWS_SECRET_ACCESS_KEY=bar
    export ENDPOINT=http://localhost:$PROXY_PORT
elif [ $CLOUD == "azblob" ]; then
    export AZURE_STORAGE_ACCOUNT=${AZURE_STORAGE_ACCOUNT:-devstoreaccount1}
    export AZURE_STORAGE_KEY=${AZURE_STORAGE_KEY:-Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==}
    export ENDPOINT=http://127.0.0.1:$PORT/$AZURE_STORAGE_ACCOUNT
    if [ ${AZURE_STORAGE_ACCOUNT} == "devstoreaccount1" ]; then
        if ! which azurite >/dev/null; then
            echo "Azurite missing, run:" >&1
            echo "npm install -g azurite" >&1
            exit 1
        fi
        rm -Rf /tmp/azblob
        mkdir -p /tmp/azblob
        PROXY_BIN="azurite-blob -l /tmp/azblob --blobPort $PORT -s"
    fi
fi

if [ "$PROXY_BIN" != "" ]; then
    $PROXY_BIN &
    PROXY_PID=$!
    export EMULATOR=1
    until curl -s "$ENDPOINT" > /dev/null; do
        echo "Waiting for proxy up..."
        sleep 1
    done
elif [ "$TIMEOUT" == "10m" ]; then
    # higher timeout for testing to real cloud
    TIMEOUT=45m
fi
