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

p=`dirname $0`
FS_BIN="${FS_BIN:-"$p/../../tigrisfs"}"

_mount() {
  local MNT_DIR=$1
  shift
  echo "=== Mount $MNT_DIR"
  "$FS_BIN" \
  --endpoint="$ENDPOINT" \
  --enable-mtime \
  --cluster \
  "$@" \
  "$BUCKET_NAME" \
  "$MNT_DIR" &
}

_umount() {
  local MNT_DIR=$1
  echo "=== Unmount $MNT_DIR"
  umount "$MNT_DIR"
  sleep 1
  until [[ $(ps -ef | grep "tigrisfs" | grep "$MNT_DIR" | wc -l) == 0 ]]; do
    echo "=== Unmount $MNT_DIR... still doing"
    sleep 1
  done
}
