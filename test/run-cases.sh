#!/bin/bash

set -ex

MNT_DIR=$(mktemp -d)
export MNT_DIR

export BUCKET_NAME="tigrisfs-test.$RANDOM"
export ENDPOINT=${ENDPOINT:-"http://localhost:8080"}

. "$(dirname "$0")/mount.sh"

_s3cmd mb s3://$BUCKET_NAME

export DEF_MNT_PARAMS="--enable-mtime --enable-specials --enable-perms"
# shellcheck disable=SC2086
_mount "$MNT_DIR" $DEF_MNT_PARAMS
trap '_umount $MNT_DIR' EXIT

sleep 5

for c in $(find test/cases -type f -name "*.sh"); do
    echo "Running $c"
    /bin/bash "$c"
done
