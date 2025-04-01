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

In first release we worked on improving reliability of the code base:
 * Removed bundled, outdated AWS SDK with critical vulnerabilities.
 * Upgraded dependencies to fix security vulnerabilities.
 * Enabled race detector in tests and fixed all detected race conditions.
 * Enabled unit, xfs and cluster tests on every PR and push to main.

# Installation

## Prebuilt DEB and RPM packages 

* Download the latest release: [DEB](https://github.com/tigrisdata/tigrisfs/releases/latest/download/tigrisfs-linux-amd64.deb), [RPM](https://github.com/tigrisdata/tigrisfs/releases/latest/download/tigrisfs-linux-amd64.rpm).
* Install the package:
  * Debian-based systems:
    ```bash
    dpkg -i tigrisfs-linux-amd64.deb
    ```
  * RPM-based systems:
    ```bash
    rpm -i tigrisfs-linux-amd64.rpm
    ```
* Configure credentials
  TigrisFS can use credentials from different sources:
  * Standard AWS credentials files `~/.aws/credentials` and `~/.aws/config`. Use `aws configure` to set them up and export
    `AWS_PROFILE` environment variable to use a specific profile.
  * Environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`.
  * TigrisFS credentials in `/etc/default/tigrisfs` or mount specific credentials in `/etc/default/tigrisfs-<bucket>`.
 
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
 
# License

Licensed under the Apache License, Version 2.0

See `LICENSE` and `AUTHORS`