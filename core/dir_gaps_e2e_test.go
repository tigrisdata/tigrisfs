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
	"fmt"
	"sync/atomic"
	"syscall"
	"time"

	. "gopkg.in/check.v1"
)

// A writer that keeps creating files whose names sort after every key on the
// server used to make every lookup of the next name list from the server: the
// listing found nothing past the name, and that was never remembered.
func (s *GoofysTest) TestLookupPastLastKeyDoesNotRelist(t *C) {
	s.fs.flags.StatCacheTTL = time.Minute
	root := s.getRoot(t)

	s.setupBlobs(s.cloud, t, map[string]*string{
		"gapdir/file-0001": nil,
		"gapdir/file-0002": nil,
		"gapdir/file-0003": nil,
	})

	var lists, heads int64
	s3 := root.fs.getCloud()
	cloud := NewTestBackend(&TestBackend{StorageBackend: s3})
	cloud.ListBlobsFunc = func(p *ListBlobsInput) (*ListBlobsOutput, error) {
		atomic.AddInt64(&lists, 1)
		return s3.ListBlobs(p)
	}
	cloud.HeadBlobFunc = func(p *HeadBlobInput) (*HeadBlobOutput, error) {
		atomic.AddInt64(&heads, 1)
		return s3.HeadBlob(p)
	}
	root.fs.setCloud(cloud)

	// Force the directory to be consulted from the server.
	root.mu.Lock()
	root.dir.DirTime = time.Time{}
	root.mu.Unlock()
	dir, err := root.LookUpCached("gapdir")
	t.Assert(err, IsNil)
	dir.mu.Lock()
	dir.dir.DirTime = time.Time{}
	dir.mu.Unlock()

	// The first unknown name past the last key may have to list once.
	_, err = dir.LookUpCached("file-0100")
	t.Assert(err, Equals, syscall.ENOENT)
	listsAfterFirst := atomic.LoadInt64(&lists)
	headsAfterFirst := atomic.LoadInt64(&heads)

	// Every later name is covered by that listing: nothing follows it.
	for i := 101; i < 110; i++ {
		_, err = dir.LookUpCached(fmt.Sprintf("file-%04d", i))
		t.Assert(err, Equals, syscall.ENOENT)
	}
	t.Assert(atomic.LoadInt64(&lists), Equals, listsAfterFirst,
		Commentf("later names must not list again; lists went %d -> %d", listsAfterFirst, atomic.LoadInt64(&lists)))
	t.Assert(atomic.LoadInt64(&heads), Equals, headsAfterFirst,
		Commentf("later names must not HEAD; heads went %d -> %d", headsAfterFirst, atomic.LoadInt64(&heads)))
}
