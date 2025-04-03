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

# Test creates and removes files and directories in random order
#

set -x

. `dirname $0`/common.sh

_s3_setup() {
  :
}

NUM_ITER="${NUM_ITER:-50}"
export GORACE="halt_on_error=1"

_cluster_setup() {
  mkdir -p "$TEST_ARTIFACTS/test_random"
  touch "$TEST_ARTIFACTS/test_random/log1" "$TEST_ARTIFACTS/test_random/log2" "$TEST_ARTIFACTS/test_random/log3"

  #opts="--debug_fuse --debug_grpc --log-format console --log-level debug --no-log-color"
  opts="--log-format console --log-level info --no-log-color"
  nodes="--cluster-peer=1:localhost:1337 --cluster-peer=2:localhost:1338 --cluster-peer=3:localhost:1339"

  MNT1=$(mktemp --suffix .node1 -d)
  _mount "$MNT1" $opts -f --log-file="$TEST_ARTIFACTS/test_random/log1" --pprof=6060 --cluster-me=1:localhost:1337 $nodes

  MNT2=$(mktemp --suffix .node2 -d)
  _mount "$MNT2" $opts -f --log-file="$TEST_ARTIFACTS/test_random/log2" --pprof=6070 --cluster-me=2:localhost:1338 $nodes

  MNT3=$(mktemp --suffix .node3 -d)
  _mount "$MNT3" $opts -f --log-file="$TEST_ARTIFACTS/test_random/log3" --pprof=6080 --cluster-me=3:localhost:1339 $nodes

  rm -rf "${MNT1:?}/*"
  rm -rf "${MNT2:?}/*"
  rm -rf "${MNT3:?}/*"

  VALID_DIR=$(mktemp -d)
  echo "=== VALID_DIR=$VALID_DIR"
}

_cleanup() {
  _umount "$MNT3"
  _umount "$MNT2"
  _umount "$MNT1"
  cat "$TEST_ARTIFACTS/test_random/log1"
  cat "$TEST_ARTIFACTS/test_random/log2"
  cat "$TEST_ARTIFACTS/test_random/log3"
}

_test() {
  sleep 5

  ps -ef|grep geese
  ps -ef|grep s3proxy

  for I in $(seq 0 "$NUM_ITER"); do
    echo "=== Iteration $I"
    MNT=$(echo -e "$MNT1\n$MNT2\n$MNT3" | shuf -n 1)

    # action with random file
    FILE=$(cd "$VALID_DIR"; find -type f | shuf -n 1)
    if [[ $FILE != "" ]]; then
      case "$((RANDOM%2))" in
      0)
        rm "$VALID_DIR/$FILE" "$MNT/$FILE"
      ;;
      1)
        dd if=/dev/urandom bs=5M count=4 | tee "$VALID_DIR/$FILE" "$MNT/$FILE" > /dev/null
        ls -alh "$VALID_DIR/$FILE"
        ls -alh "$MNT/$FILE"
        cat "$VALID_DIR/$FILE" > "$TEST_ARTIFACTS/test_random/valid_file"
        cat "$MNT/$FILE" > "$TEST_ARTIFACTS/test_random/mnt_file"
        cmp "$TEST_ARTIFACTS/test_random/valid_file" "$TEST_ARTIFACTS/test_random/mnt_file"
      ;;
      esac
    fi

    # action with random dir
    FILE=$(cd "$VALID_DIR"; find -type d | shuf -n 1)
    if [ -d "$VALID_DIR/$FILE" ]; then
        case "$((RANDOM%7))" in
        0)
          [[ "$FILE" == '.' ]] || rm -rf "$VALID_DIR/$FILE" "$MNT/$FILE"
        ;;
        1|2|3|4)
          local NEW_DIR
          NEW_DIR="$(echo -e "a\nb\n\c\nd" | shuf -n 1)"
          mkdir -p "$VALID_DIR/$FILE/$NEW_DIR" "$MNT/$FILE/$NEW_DIR"
        ;;
        5|6)
          for J in $(seq $((RANDOM%50))); do
            RAND_PREFIX=$((RANDOM%2))
            case $((RANDOM%2)) in
            0)
              dd if=/dev/urandom bs=5M count=2 | tee "$VALID_DIR/$FILE/${RAND_PREFIX}_$J" "$MNT/$FILE/${RAND_PREFIX}_$J" > /dev/null
            ;;
            1)
              dd if=/dev/urandom bs=20 count=1 | tee "$VALID_DIR/$FILE/${RAND_PREFIX}_$J" "$MNT/$FILE/${RAND_PREFIX}_$J" > /dev/null
            ;;
            esac
            ls -alh "$VALID_DIR/$FILE/${RAND_PREFIX}_$J"
            ls -alh "$MNT/$FILE/${RAND_PREFIX}_$J"
            cat "$VALID_DIR/$FILE/${RAND_PREFIX}_$J" > "$TEST_ARTIFACTS/test_random/valid_file"
            cat "$MNT/$FILE/${RAND_PREFIX}_$J" > "$TEST_ARTIFACTS/test_random/mnt_file"
            cmp "$TEST_ARTIFACTS/test_random/valid_file" "$TEST_ARTIFACTS/test_random/mnt_file"
          done
        ;;
        esac
    fi

    #tree "$VALID_DIR"

    diff -y <(cd "$VALID_DIR"; find . | sort) <(cd "$MNT1"; find . | fgrep -v '-' | sort)
    diff -y <(cd "$MNT1"; find . | sort) <(cd "$MNT2"; find . | sort)
    diff -y <(cd "$MNT2"; find . | sort) <(cd "$MNT3"; find . | sort)
  done
}

_s3_validate() {
  for FILE in $(cd "$VALID_DIR"; find -type f | cut -c 3- | grep .); do
    [ -d "$VALID_DIR/$FILE" ] || diff <(_s3cmd get "s3://test/$FILE" -) "$VALID_DIR/$FILE"
  done
}

_check
