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

set -ex

. "$(dirname "$0")/../mount.sh"

CASE=list_content

DIR="$MNT_DIR/$CASE/$RANDOM"
mkdir -p "$DIR"
fn="$DIR/file"

NUM_FILES=999

set +x
for i in $(seq $NUM_FILES); do
  echo "test $i" > "$fn.$i"
done
set -x

# Drop caches by remounting
_umount "$MNT_DIR"
# shellcheck disable=SC2086
FS_BIN=$(dirname "$0")/../../tigrisfs _mount "$MNT_DIR" $DEF_MNT_PARAMS --tigris-list-content
sleep 5

start=$(date +%s%3N)
# Test list content
ls -la "$DIR" | grep -E "file.[0-9]{1,3}" | wc -l | grep $NUM_FILES

set +x
for i in $(seq $NUM_FILES); do
  grep -e "test $i" "$fn.$i" || exit 1
done
set -x
end=$(date +%s%3N)
echo "Duration: $((end - start)) ms"

# Create new set of files
DIR="$MNT_DIR/$CASE/no-preload/$RANDOM"
mkdir -p "$DIR"
fn="$DIR/file"

set +x
for i in $(seq $NUM_FILES); do
  echo "test $i" > "$fn.$i"
done
set -x

# Drop caches by remounting and not set --tigris-list-content
_umount "$MNT_DIR"
# shellcheck disable=SC2086
FS_BIN=$(dirname "$0")/../../tigrisfs _mount "$MNT_DIR" $DEF_MNT_PARAMS
sleep 5

start=$(date +%s%3N)
# Test list content
ls -la "$DIR" | grep -E "file.[0-9]{1,3}" | wc -l | grep $NUM_FILES

set +x
for i in $(seq $NUM_FILES); do
  grep -e "test $i" "$fn.$i" || exit 1
done
set -x
end=$(date +%s%3N)
echo "Duration: $((end - start)) ms"
