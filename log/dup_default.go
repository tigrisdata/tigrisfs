//go:build !windows

// Copyright 2021 Yandex LLC
// Copyright 2024 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"io"
	"log/syslog"
	"os"

	"golang.org/x/sys/unix"
)

func redirectStdout(target *os.File) error {
	return unix.Dup2(int(target.Fd()), int(os.Stdout.Fd()))
}

func redirectStderr(target *os.File) error {
	return unix.Dup2(int(target.Fd()), int(os.Stderr.Fd()))
}

func InitSyslog() (io.Writer, error) {
	return syslog.New(syslog.LOG_INFO, "tigrisfs")
}
