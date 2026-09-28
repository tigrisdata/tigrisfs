// Copyright 2026 Tigris Data, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"syscall"

	"github.com/tigrisdata/tigrisfs/core/cfg"
	. "gopkg.in/check.v1"
)

type LoadRangeTest struct{}

var _ = Suite(&LoadRangeTest{})

// inodeAheadOfServer builds an inode whose local size has run ahead of the size
// the server knows, which is the ordinary state of a file being appended to
// before its flush lands. No buffers are held, so every read is a hole.
func inodeAheadOfServer(localSize, knownSize uint64) *Inode {
	fs := &Goofys{flags: cfg.DefaultFlags()}
	inode := NewInode(fs, nil, "appending-file")
	inode.Attributes.Size = localSize
	inode.knownSize = knownSize

	return inode
}

// A hole beginning past knownSize has nothing behind it on the server. Clamping
// its end to knownSize would invert it, and the unsigned subtractions downstream
// then underflow: splitRA used to panic on it and take the whole daemon down.
// LoadRange must refuse the range instead, so its callers can treat the file as
// remotely changed rather than crashing.
func (s *LoadRangeTest) TestLoadRangeRefusesHolePastKnownSize(t *C) {
	inode := inodeAheadOfServer(8*1024*1024, 1024*1024)
	inode.mu.Lock()
	defer inode.mu.Unlock()

	miss, err := inode.LoadRange(4*1024*1024, 1024*1024, 0, true)
	t.Assert(err, Equals, syscall.ERANGE)
	t.Assert(miss, Equals, true)

	// The callers that reset the inode cache key off exactly this mapping, so a
	// change that stopped mapping to ERANGE would silently skip that handling.
	t.Assert(mapAwsError(err), Equals, syscall.ERANGE)
}
