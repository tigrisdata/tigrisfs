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

// Just a check.v1 wrapper to allow running selected tests with:
// go test -v internal_test.go lfru_btree_test.go lfru_btree.go

package core

import (
	"os"
	"testing"

	"github.com/tigrisdata/tigrisfs/core/cfg"
	"github.com/tigrisdata/tigrisfs/log"
	. "gopkg.in/check.v1"
)

var testLog = log.GetLogger("test")

func TestCheckSuites(t *testing.T) {
	TestingT(t)
}

func TestMain(m *testing.M) {
	cfg.InitLoggers(&cfg.FlagStorage{LogLevel: "warn", Foreground: true, LogFormat: "console"})

	log.DumpLoggers("TestMain")

	os.Exit(m.Run())
}
