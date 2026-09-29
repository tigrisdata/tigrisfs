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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tigrisdata/tigrisfs/core/cfg"
)

func cfgDefaultFlagsForTest() *cfg.FlagStorage { return cfg.DefaultFlags() }

func gapRanges(d *DirInodeData) [][2]string {
	out := make([][2]string, 0, len(d.Gaps))
	for _, g := range d.Gaps {
		out = append(out, [2]string{g.start, g.end})
	}
	return out
}

// The slurp cache was inert: markGapLoaded's final-length arithmetic was one
// short, so the range it inserted was truncated away every time. With Gaps
// always empty, every lookup on an expired directory went back to the server
// for a listing and a HEAD.
func TestMarkGapLoadedKeepsWhatItInserts(t *testing.T) {
	old := time.Now().Add(-time.Hour)
	d := &DirInodeData{}

	d.markGapLoaded("a", "m")
	require.Equal(t, [][2]string{{"a", "m"}}, gapRanges(d))
	require.True(t, d.checkGapLoaded("b", old))
	require.False(t, d.checkGapLoaded("a", old), "start is exclusive: StartAfter never returns it")
	require.True(t, d.checkGapLoaded("m", old), "end is inclusive")
	require.False(t, d.checkGapLoaded("z", old))

	d.markGapLoaded("p", "s")
	require.Equal(t, [][2]string{{"a", "m"}, {"p", "s"}}, gapRanges(d), "append after")

	d.markGapLoaded("0", "5")
	require.Equal(t, [][2]string{{"0", "5"}, {"a", "m"}, {"p", "s"}}, gapRanges(d), "insert before")
}

func TestMarkGapLoadedOverlaps(t *testing.T) {
	d := &DirInodeData{}
	d.markGapLoaded("a", "m")
	d.markGapLoaded("p", "s")

	d.markGapLoaded("k", "q")
	require.Equal(t, [][2]string{{"a", "k"}, {"k", "q"}, {"q", "s"}}, gapRanges(d),
		"overlapping neighbours are trimmed to the parts outside the new range")

	d.markGapLoaded("b", "r")
	require.Equal(t, [][2]string{{"a", "b"}, {"b", "r"}, {"r", "s"}}, gapRanges(d),
		"a new range covering several old ones replaces their interiors")

	d.markGapLoaded("", gapEndOfListing)
	require.Equal(t, [][2]string{{"", gapEndOfListing}}, gapRanges(d), "a full listing subsumes everything")
}

// A narrow re-list inside an open-ended range must not throw away what is
// known past its end, or the next lookup of a later name lists again.
func TestMarkGapLoadedSplitsAContainingRange(t *testing.T) {
	old := time.Now().Add(-time.Hour)
	d := &DirInodeData{}
	d.markGapLoaded("a", gapEndOfListing)

	d.markGapLoaded("f", "h")
	require.Equal(t, [][2]string{{"a", "f"}, {"f", "h"}, {"h", gapEndOfListing}}, gapRanges(d))
	require.True(t, d.checkGapLoaded("zzz", old), "still known that nothing follows h")
}

// This is the reported symptom: an appending writer creates names sorting
// after every key on the server. A listing from such a name returns nothing,
// and must be remembered as "nothing after this", not as an empty range.
func TestGapRecordsEndOfListing(t *testing.T) {
	old := time.Now().Add(-time.Hour)
	d := &DirInodeData{}

	// What the slurp records when the listing is not truncated.
	d.markGapLoaded("dir/file-0100", gapEndOfListing)

	for _, later := range []string{"dir/file-0101", "dir/file-0101/", "dir/file-9999", "zzz"} {
		require.True(t, d.checkGapLoaded(later, old), "%q sorts after the last listed key", later)
	}
	require.False(t, d.checkGapLoaded("dir/file-0099", old), "earlier names were not listed")

	// Keys are UTF-8, so the sentinel must sort after any of them.
	require.True(t, d.checkGapLoaded("\U0010FFFF", old))
	require.True(t, d.checkGapLoaded("\xf4\x8f\xbf\xbf", old))
}

func TestCheckGapLoadedDropsStaleRanges(t *testing.T) {
	d := &DirInodeData{}
	d.markGapLoaded("a", "m")
	d.Gaps[0].loadTime = time.Now().Add(-2 * time.Minute)

	require.False(t, d.checkGapLoaded("b", time.Now().Add(-time.Minute)), "older than the TTL")
	require.Empty(t, d.Gaps, "a stale range is evicted on lookup")
}

// An explicit refresh must forget the loaded ranges too, or LookUp keeps
// serving the entries the refresh just expired straight from cache.
func TestResetDirTimeDropsLoadedRanges(t *testing.T) {
	fs := &Goofys{flags: cfgDefaultFlagsForTest()}
	root := NewInode(fs, nil, "")
	root.ToDir()
	sub := NewInode(fs, root, "sub")
	sub.ToDir()
	file := NewInode(fs, sub, "file")
	root.dir.markGapLoaded("", gapEndOfListing)
	require.NotEmpty(t, root.dir.Gaps)

	// resetDirTimeRec calls this first; the rest of it needs a mounted fs, so
	// the wiring is covered by the notify-refresh tests over FUSE.
	file.dropLoadedRanges()
	require.Empty(t, root.dir.Gaps, "refreshing any inode drops the root's ranges")
}
