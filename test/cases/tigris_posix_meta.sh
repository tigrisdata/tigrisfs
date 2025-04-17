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

CASE=posix_meta
DIR="$MNT_DIR/$CASE/$RANDOM"

mkdir -p "$DIR"
fn="$DIR/file.$RANDOM"

# Test group id
touch "$fn"
[ "$(stat -c '%G' $fn)" == "$USER" ] || exit 1
chown "$USER:users" "$fn"
[ "$(stat -c '%G' "$fn")" == "users" ] || exit 1
touch "$fn"
[ "$(stat -c '%G' $fn)" == "users" ] || exit 1
[ "$(stat -c '%A' $fn)" == "-rw-rw-r--" ] || exit 1
chmod 600 "$fn"
[ "$(stat -c '%A' $fn)" == "-rw-------" ] || exit 1

# Test symlink
ln -s "$fn" "${fn}_link"
[ -L "${fn}_link" ] || exit 1

# Test fifo
mkfifo "${fn}_fifo"
ls -la $DIR
[ "$(stat -c '%A' ${fn}_fifo)" == "prw-rw-r--" ] || exit 1
chmod 600 "${fn}_fifo"
[ "$(stat -c '%A' ${fn}_fifo)" == "prw-------" ] || exit 1
[ -p "${fn}_fifo" ] || exit 1

# Test attributes persisted after restart
_umount "$MNT_DIR"
# shellcheck disable=SC2086
FS_BIN=$(dirname "$0")/../../tigrisfs _mount "$MNT_DIR" $DEF_MNT_PARAMS
sleep 5

[ "$(stat -c '%G' $fn)" == "users" ] || exit 1
[ "$(stat -c '%A' $fn)" == "-rw-------" ] || exit 1
[ -L "${fn}_link" ] || exit 1
[ -p "${fn}_fifo" ] || exit 1
[ "$(stat -c '%A' ${fn}_fifo)" == "prw-------" ] || exit 1
