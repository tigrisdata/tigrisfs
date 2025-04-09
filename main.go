// Copyright 2015 - 2017 Ka-Hing Cheung
// Copyright 2015 - 2017 Google Inc. All Rights Reserved.
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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/tigrisdata/tigrisfs/core"
	"github.com/tigrisdata/tigrisfs/core/cfg"
	"github.com/tigrisdata/tigrisfs/log"
	"github.com/urfave/cli"

	_ "net/http/pprof"
)

var mainLog = log.GetLogger("main")

func registerSIGINTHandler(fs *core.Goofys, mfs core.MountedFS, flags *cfg.FlagStorage) {
	// Register for SIGINT.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signalsToHandle...)

	// Start a goroutine that will unmount when the signal is received.
	go func() {
		for {
			s := <-signalChan
			if isSigUsr1(s) {
				mainLog.Info().Str("signal", fmt.Sprintf("%v", s)).Msg("Received signal")
				fs.SigUsr1()
				continue
			}

			mainLog.Info().Str("signal", fmt.Sprintf("%v", s)).Msg("Received signal, attempting to unmount...")

			err := mfs.Unmount()
			if err != nil {
				mainLog.Error().Str("signal", fmt.Sprintf("%v", s)).Err(err).Msg("Failed to unmount in response to signal")
			} else {
				mainLog.Info().Str("mountPoint", flags.MountPoint).Str("signal", fmt.Sprintf("%v", s)).Msg("Successfully unmounted in response to signal")
				return
			}
		}
	}()
}

func main() {
	messagePath()

	app := cfg.NewApp()

	var flags *cfg.FlagStorage
	var child *os.Process

	app.Action = func(c *cli.Context) (err error) {
		// We should get two arguments exactly. Otherwise error out.
		if len(c.Args()) != 2 {
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Error: %s takes exactly two arguments.\n\n",
				app.Name)
			mainLog.E(cli.ShowAppHelp(c))
			os.Exit(1)
		}

		// Populate and parse flags.
		bucketName := c.Args()[0]
		flags = cfg.PopulateFlags(c)
		if flags == nil {
			mainLog.E(cli.ShowAppHelp(c))
			err = fmt.Errorf("invalid arguments")
			return
		}
		defer func() {
			time.Sleep(time.Second)
			flags.Cleanup()
		}()

		var daemonizer *Daemonizer
		if !canDaemonize {
			flags.Foreground = true
		}

		if !flags.Foreground {
			daemonizer = NewDaemonizer()
			// Do not close stderr before mounting to print mount errors
			initLogFile := flags.LogFile
			if initLogFile == "" || initLogFile == "syslog" {
				initLogFile = "stderr"
			}
			if err = daemonizer.Daemonize(initLogFile); err != nil {
				return err
			}
		}

		cfg.InitLoggers(flags)

		mainLog.Info().Str("version", cfg.Version).Msg("Starting TigrisFS version")
		mainLog.Info().Str("bucketName", bucketName).Str("mountPoint", flags.MountPoint).Msg("Mounting")
		mainLog.Info().Uint32("uid", flags.Uid).Uint32("defaultGid", flags.Gid).Msg("Default uid=gid")

		// Mount the file system.
		fs, mfs, err := mount(
			context.Background(),
			bucketName,
			flags)

		pprof := flags.PProf
		if pprof == "" && os.Getenv("PPROF") != "" {
			pprof = os.Getenv("PPROF")
		}
		if pprof != "" {
			go func() {
				addr := pprof
				if !strings.Contains(addr, ":") {
					addr = "127.0.0.1:" + addr
				}
				mainLog.Println(http.ListenAndServe(addr, nil))
			}()
		}

		if err != nil {
			if !flags.Foreground {
				daemonizer.NotifySuccess(false)
			}
			mainLog.Fatal().Err(err).Msg("Mounting file system failed")
			// fatal also terminates itself
		} else {
			mainLog.Info().Msg("File system has been successfully mounted.")
			if !flags.Foreground {
				daemonizer.NotifySuccess(true)
				mainLog.Debug().Msg("Process is daemonized. Closing stdout and stderr.")
				mainLog.E(os.Stderr.Close())
				mainLog.E(os.Stdout.Close())
			}
			// Let the user unmount with Ctrl-C (SIGINT)
			registerSIGINTHandler(fs, mfs, flags)

			// Drop root privileges
			if flags.Setuid != 0 {
				mainLog.E(setuid(flags.Setuid))
			}
			if flags.Setgid != 0 {
				mainLog.E(setgid(flags.Setgid))
			}

			// Wait for the file system to be unmounted.
			err = mfs.Join(context.Background())
			if err != nil {
				err = fmt.Errorf("MountedFileSystem.Join: %v", err)
				return
			}
			fs.SyncTree(nil)

			mainLog.Info().Msg("Successfully exiting.")
		}
		return
	}

	err := app.Run(cfg.MessageMountFlags(os.Args))
	if err != nil {
		if flags != nil && !flags.Foreground && child != nil {
			mainLog.Fatal().Msg("Unable to mount file system, see syslog for details")
		}
		os.Exit(1)
	}
}
