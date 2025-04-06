// Copyright 2015 - 2017 Ka-Hing Cheung
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
	"fmt"
	"io"
	glog "log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var DefaultLogConfig = &LogConfig{
	Level:  "info",
	Format: "console",
	Color:  false,
}

var (
	mu      sync.Mutex
	loggers = make(map[string]*LogHandle)
)

var logWriter io.Writer = os.Stderr

func logStderr(msg string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, msg, args...)
}

func InitLoggerRedirect(logFileName string) {
	if logFileName == "syslog" {
		logWriter = InitSyslog()
	} else if logFileName != "stderr" && logFileName != "/dev/stderr" && logFileName != "" {
		var err error
		lf, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			logStderr("Couldn't open file %v for writing logs", logFileName)
			return
		}
		if err = redirectStdout(lf); err != nil {
			logStderr("Couldn't redirect STDOUT to the log file %v", logFileName)
			return
		}
		if err = redirectStderr(lf); err != nil {
			logStderr("Couldn't redirect STDERR to the log file %v", logFileName)
			return
		}
		logWriter = lf
	}
}

func SetCloudLogLevel(level zerolog.Level) {
	for _, logr := range loggers {
		//		if k != "main" && k != "fuse" {
		logr.Level(level)
		//		}
	}

	DumpLoggers("SetCloudLogLevel")
}

func SetLoggersConfig(config *LogConfig) {
	mu.Lock()
	defer mu.Unlock()

	for k, l := range loggers {
		nl := NewLogger(config, l.name, config.Color, logWriter)
		loggers[k].Logger = nl.Logger
	}
}

type LogHandle struct {
	//	logrus.Logger
	*zerolog.Logger

	name string
}

// for aws.Logger
func (l *LogHandle) Log(args ...interface{}) {
	l.Debug().CallerSkipFrame(1).Msgf("%+v", args...)
}

func (l *LogHandle) Infof(msg string, args ...interface{}) {
	l.Info().CallerSkipFrame(1).Msgf(msg, args...)
}

func (l *LogHandle) Errorf(msg string, args ...interface{}) {
	l.Error().CallerSkipFrame(1).Msgf(msg, args...)
}

func (l *LogHandle) Warnf(msg string, args ...interface{}) {
	l.Warn().CallerSkipFrame(4).Msgf(msg, args...)
}

func (l *LogHandle) Debugf(msg string, args ...interface{}) {
	l.Debug().CallerSkipFrame(1).Msgf(msg, args...)
}

func (l *LogHandle) IsLevelEnabled(level zerolog.Level) bool {
	return l.GetLevel() <= level
}

func (l *LogHandle) SetLevel(level zerolog.Level) {
	*l.Logger = l.Level(level)
}

func (l *LogHandle) E(err error) bool {
	if err == nil {
		return false
	}

	l.Error().CallerSkipFrame(1).Msg(err.Error())

	return true
}

func GetLogger(name string) *LogHandle {
	mu.Lock()
	defer mu.Unlock()

	logger, ok := loggers[name]
	if !ok {
		logger = NewLogger(DefaultLogConfig, name, DefaultLogConfig.Color, logWriter)
		loggers[name] = logger
	}

	return logger
}

func GetStdLogger(l *zerolog.Logger) *glog.Logger {
	return glog.New(l, "", 0)
}

type LogConfig struct {
	Level      string
	Format     string
	Color      bool
	SampleRate float64 `json:"sample_rate" mapstructure:"sample_rate" yaml:"sample_rate"`
}

func consoleFormatCallerWithModule(i any, module string) string {
	var c string
	if cc, ok := i.(string); ok {
		c = cc
	}
	if len(c) > 0 {
		l := strings.Split(c, "/")
		if len(l) == 1 {
			return l[0]
		}
		return l[len(l)-2] + "/" + l[len(l)-1]
	}
	return module + " " + c
}

func NewLogger(config *LogConfig, module string, colorized bool, writer io.Writer) *LogHandle {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lvl, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error parsing log level. defaulting to info level")
		lvl = zerolog.InfoLevel
	}

	var logger zerolog.Logger
	if config.Format == "console" {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.StampMicro,
		}
		output.NoColor = !colorized
		output.FormatCaller = func(i any) string {
			return consoleFormatCallerWithModule(i, module)
		}
		logger = zerolog.New(output).Level(lvl).With().Timestamp().CallerWithSkipFrameCount(2).Stack().Logger()
	} else {
		logger = zerolog.New(writer).Level(lvl).With().Timestamp().CallerWithSkipFrameCount(2).Stack().
			Str("module", module).Logger()
	}

	return &LogHandle{Logger: &logger}
}

func DumpLoggers(name string) {
	for k, l := range loggers {
		fmt.Printf("%v Logger %v: %v\n", name, k, l.GetLevel().String())
	}
}

/*
// E is a helper function to shortcut condition checking and logging
// in the case of error
// Used like this:
//
//	if E(err) {
//	    return err
//	}
//
// to replace:
//
//	if err != nil {
//	    log.Msgf(err.Error())
//	    return err
//	}
func E(err error) bool {
	if err == nil {
		return false
	}

	log.Error().CallerSkipFrame(1).Msg(err.Error())

	return true
}

// CE is a helper to shortcut error creation and logging
// Used like this:
//
// return CE("msg, value %v", value)
//
// to replace:
//
// err := fmt.Errorf("msg, value %v", value)
// log.Msgf("msg, value %v", value)
// return err.
func CE(format string, args ...any) error {
	err := fmt.Errorf(format, args...)

	log.Error().CallerSkipFrame(1).Msg(err.Error())

	return err
}
*/
