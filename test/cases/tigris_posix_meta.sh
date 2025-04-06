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
DIR=$MNT_DIR/$CASE/$RANDOM

df -h

mkdir -p $DIR
fn=$DIR/file.$RANDOM
touch $fn
[ "$(stat -c '%G' $fn)" == "$USER" ]
chown $USER:users $fn
[ "$(stat -c '%G' $fn)" == "users" ]
touch $fn
[ "$(stat -c '%G' $fn)" == "users" ]

_umount $MNT_DIR
FS_BIN=$(dirname "$0")/../../tigrisfs _mount $MNT_DIR --enable-perms
sleep 5

[ "$(stat -c '%G' $fn)" == "users" ]
