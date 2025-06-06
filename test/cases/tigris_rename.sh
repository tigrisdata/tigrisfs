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

set -exo pipefail

. "$(dirname "$0")/../mount.sh"

CASE=tigris_rename
KEY="$CASE/$RANDOM"
DIR="$MNT_DIR/$KEY"
TMP_DIR=/tmp/$KEY
size=32768
n_iter=1

export AWS_PROFILE=local

mount() {
	# shellcheck disable=SC2086
	FS_BIN=$(dirname "$0")/../../tigrisfs _mount $@
	sleep 5
}

test_copy() {
	mount "$MNT_DIR" --no-instant-rename --debug_s3 --log-file=./test-tigris-copy.log --log-format=console --no-log-color

	mkdir -p "$DIR"
	mkdir -p "$TMP_DIR"

	n=$RANDOM
	for i in $(seq 1 $n_iter); do
	  fn="$DIR/file.$n.$i"
	  fntmp="$TMP_DIR/file.$n.$i"
	  head -c $size < /dev/urandom > "$fntmp"
	  cp "$fntmp" "$fn"
	done

	start=$(date +%s%3N)

	for i in $(seq 1 $n_iter); do
	  fn="$DIR/file.$n.$i"
	  mv "$fn" "$fn.tmp"
	done

	sync

	end=$(date +%s%3N)
	echo "Duration0: $((end - start)) ms"

	#for i in $(seq 1 10); do
	#  fn="$DIR/file.$i"
	#  cat "$fn.$i.tmp" > /dev/null
	#done

	# Remount without specials enabled
	_umount "$MNT_DIR"

	end=$(date +%s%3N)
	echo "Duration: $((end - start)) ms"

	for i in $(seq 1 $n_iter); do
	  fn="file.$n.$i"
	  aws s3api get-object --key "$KEY/$fn.tmp" --bucket "$BUCKET_NAME" "$TMP_DIR/$fn.tmp"
	  diff -u "$TMP_DIR/$fn.tmp" "$TMP_DIR/$fn"
	done
}

test_rename_one() {
	mount "$MNT_DIR" --debug_s3 --log-file=./test-tigris-rename.log --log-format=console --no-log-color

	mkdir -p "$DIR"
	mkdir -p "$TMP_DIR"

	n=$RANDOM
	for i in $(seq 1 $n_iter); do
	  fn="$DIR/file.$n.$i"
	  fntmp="$TMP_DIR/file.$n.$i"
	  head -c $size < /dev/urandom > "${fntmp}"
	  cp "$fntmp" "$fn"
	done

	start=$(date +%s%3N)

	for i in $(seq 1 $n_iter); do
	  fn="$DIR/file.$n.$i"
	  mv "${fn}" "${fn}.tmp"
	done

	sync

	end=$(date +%s%3N)
	echo "Duration0: $((end - start)) ms"

	_umount "$MNT_DIR"

	for i in $(seq 1 $n_iter); do
	  fn="file.$n.$i"
	  aws s3api get-object --key "$KEY/$fn.tmp" --bucket "$BUCKET_NAME" "$TMP_DIR/$fn.tmp"
	  diff -u "$TMP_DIR/$fn.tmp" "$TMP_DIR/$fn"
	done

	end=$(date +%s%3N)
	echo "Duration: $((end - start)) ms"
}

test_rename_nested() {
	mount "$MNT_DIR" --debug_fuse --pprof 8889 --debug_s3 --log-file=./test-tigris-rename.log --log-format=console --no-log-color

	n=$RANDOM
	DIR_NESTED="$DIR/$n/nested"
	DIR_NESTED_2="$DIR_NESTED/nested2"
	TMP_DIR_NESTED="$TMP_DIR/$n/nested"
	TMP_DIR_NESTED_2="$TMP_DIR/$n/nested/nested2"

	mkdir -p "$DIR_NESTED" "$DIR_NESTED_2"
	mkdir -p "$TMP_DIR_NESTED" "$TMP_DIR_NESTED_2"

	for i in $(seq 1 $n_iter); do
	  fn="$DIR_NESTED/file.$n.$i"
	  fntmp="$TMP_DIR_NESTED/file.$n.$i"
	  head -c $size < /dev/urandom > "${fntmp}"
	  cp "$fntmp" "$fn"
	done

	for i in $(seq 1 $n_iter); do
	  fn="$DIR_NESTED_2/file.$n.$i"
	  fntmp="$TMP_DIR_NESTED_2/file.$n.$i"
	  head -c $size < /dev/urandom > "${fntmp}"
	  cp "$fntmp" "$fn"
	done

	start=$(date +%s%3N)

	sync
	sleep 5

	tree "$DIR"

	DIR_NESTED_3="$DIR/nested3"
	mv "$DIR_NESTED" "$DIR_NESTED_3"

	tree "$DIR"

	sync
	sleep 5

	end=$(date +%s%3N)
	echo "Duration0: $((end - start)) ms"

	_umount "$MNT_DIR"

	for i in $(seq 1 $n_iter); do
	  fn="$KEY/nested3/file.$n.$i"
	  fntmp="$TMP_DIR_NESTED/file.$n.$i"
	  aws s3api get-object --key "$fn" --bucket "$BUCKET_NAME" "$TMP_DIR/tmp0"
	  diff -u "$TMP_DIR/tmp0" "$fntmp"
	done

	for i in $(seq 1 $n_iter); do
	  fn="$KEY/nested3/nested2/file.$n.$i"
	  fntmp="$TMP_DIR_NESTED_2/file.$n.$i"
	  aws s3api get-object --key "$fn" --bucket "$BUCKET_NAME" "$TMP_DIR/tmp1"
	  diff -u "$TMP_DIR/tmp1" "$fntmp"
	done

	end=$(date +%s%3N)
	echo "Duration: $((end - start)) ms"
}

test_rename() {
	test_rename_one
	test_rename_nested
}

_umount "$MNT_DIR"

test_copy
test_rename

mount "$MNT_DIR" $DEF_MNT_PARAMS

