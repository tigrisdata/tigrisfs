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

import "github.com/rs/zerolog"

var defaultLog = GetLogger("default")

func Error() *zerolog.Event {
	return defaultLog.Error()
}

func Warn() *zerolog.Event {
	return defaultLog.Warn()
}

func Info() *zerolog.Event {
	return defaultLog.Info()
}

func Debug() *zerolog.Event {
	return defaultLog.Debug()
}

func Fatal() *zerolog.Event {
	return defaultLog.Fatal()
}

func Trace() *zerolog.Event {
	return defaultLog.Trace()
}
