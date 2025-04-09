#!/bin/bash

set -ex

MNT_DIR=$(mktemp -d)
export MNT_DIR

export BUCKET_NAME="tigrisfs-test.$RANDOM"
export ENDPOINT=${ENDPOINT:-"http://localhost:8080"}

. "$(dirname "$0")/mount.sh"

_s3cmd mb s3://$BUCKET_NAME

_mount "$MNT_DIR" --enable-specials --enable-perms
trap '_umount $MNT_DIR' EXIT

sleep 5

for c in $(find test/cases -type f -name "*.sh"); do
    echo "Running $c"
    /bin/bash "$c"
done
