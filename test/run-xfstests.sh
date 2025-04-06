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

set -e

if [ "$NO_PROXY" == "" ]; then
  . `dirname $0`/run-proxy.sh
fi

mkdir -p /tmp/tigrisfs
mkdir -p /tmp/tigrisfs2

sleep 5

[ -e /usr/bin/tigrisfs ] || sudo ln -s `pwd`/tigrisfs /usr/bin/tigrisfs

sudo apt-get update
sudo apt-get -y install s3cmd

s3cmd --signature-v2 --no-ssl --host-bucket= --access_key=foo --secret_key=bar --host=http://localhost:$PROXY_PORT mb s3://testbucket
s3cmd --signature-v2 --no-ssl --host-bucket= --access_key=foo --secret_key=bar --host=http://localhost:$PROXY_PORT mb s3://testbucket2

sed 's/\/home\/tigrisfs\/xfstests/$(pwd)/' <test/xfstests.config >xfstests/local.config

cp test/sync_unmount.py xfstests/

cd xfstests
sudo apt-get -y install xfslibs-dev uuid-dev libtool-bin \
	e2fsprogs automake gcc libuuid1 quota attr make \
	libacl1-dev libaio-dev xfsprogs libgdbm-dev gawk fio dbench \
	uuid-runtime python3 sqlite3 libcap-dev
make -j8
sudo -E ./check -fuse generic/001 generic/005 generic/006 generic/007 generic/011 generic/013 generic/014
