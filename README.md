[![unittests](https://github.com/tigrisdata/tigrisfs/actions/workflows/test.yaml/badge.svg)]()
[![xfstests](https://github.com/tigrisdata/tigrisfs/actions/workflows/xfstests.yaml/badge.svg)]()
[![cluster-test](https://github.com/tigrisdata/tigrisfs/actions/workflows/cluster_test.yaml/badge.svg)]()


TigrisFS is a high-performance FUSE-based file system for S3-compatible object storage written in Go.

# Overview

TigrisFS allows you to mount an S3 or compatible object store bucket as a local file system.

TigrisFS is based on [GeeseFS](https://github.com/yandex-cloud/geesefs), which is a fork of Goofys.
GeeseFS focused on solving performance problems which FUSE file systems based on S3 typically have,
especially with small files and metadata operations.
It solves these problems by using aggressive parallelism and asynchrony.

The goal of TigrisFS is to further improve on performance and reliability especially in distributed cluster setup.

In the first release we focused on improving security and reliability of the code base:
 * Improved security by:
   * removing bundled, outdated AWS SDK with critical vulnerabilities.
   * upgrading dependencies to fix security vulnerabilities.
 * Improved reliability by:
   * fixing all race conditions found by race detector, which is now enabled by default in tests.
   * fixing all linter issues and enabling linting by default in CI.
   * running more extensive tests and enabling them by default in CI.

## Tigris specific features

When mounted with the [Tigris](https://www.tigrisdata.com) backend TigrisFS supports:
  * POSIX permissions, special files, symbolic links.
  * Auto-preload content of small files on directory list in single request.
  * Allows to auto prefetch directory data to the region on list.

# Installation

## Prebuilt DEB and RPM packages 

* Download the latest release: [DEB](https://github.com/tigrisdata/tigrisfs/releases/download/v1.2.0/tigrisfs_1.2.0_linux_amd64.deb), [RPM](https://github.com/tigrisdata/tigrisfs/releases/download/v1.2.0/tigrisfs_1.2.0_linux_amd64.rpm).
* Install the package:
  * Debian-based systems:
    ```bash
    dpkg -i tigrisfs_1.2.0_linux_amd64.deb
    ```
  * RPM-based systems:
    ```bash
    rpm -i tigrisfs_1.2.0_linux_amd64.rpm
    ```
* Configure credentials
  TigrisFS can use credentials from different sources:
  * Standard AWS credentials files `~/.aws/credentials` and `~/.aws/config`. Use `aws configure` to set them up and export
    `AWS_PROFILE` environment variable to use a specific profile.
  * Environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`.
  * TigrisFS credentials in `/etc/default/tigrisfs` or mount specific credentials in `/etc/default/tigrisfs-<bucket>`.
See [docs](https://www.tigrisdata.com/docs/sdks/s3/aws-cli/) for more details.
 
* Mount the bucket
  * as current user
    ```bash
    systemctl --user start tigrisfs@<bucket>
    ```
    The bucket is mounted at `$HOME/mnt/tigrisfs/<bucket>`.
  * as root
    ```bash
    systemctl start tigrisfs@<bucket>
    ```
    The bucket is mounted at `/mnt/tigrisfs/<bucket>`.

## Binary install

* Download and unpack the latest release:
  * MacOS ARM64
    ```
      curl -L https://github.com/tigrisdata/tigrisfs/releases/download/v1.2.0/tigrisfs_1.2.0_darwin_arm64.tar.gz | sudo tar -xz -C /usr/local/bin
    ```
* Configuration is the same as for the DEB and RPM packages above.
* Mount the bucket:
  * as current user
    ```bash
    /usr/local/bin/tigrisfs <bucket> $HOME/mnt/tigrisfs/<bucket>
    ```
  * as root
    ```bash
    sudo /usr/local/bin/tigrisfs <bucket> /mnt/tigrisfs/<bucket>
    ```
 
# License

Licensed under the Apache License, Version 2.0

See `LICENSE` and `AUTHORS`