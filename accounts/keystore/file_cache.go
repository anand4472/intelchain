// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package keystore

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"

	"github.com/zennittians/intelchain/internal/utils"
)

// fileCache is a cache of files seen during scan of keystore.
type fileCache struct {
	all     mapset.Set // Set of all files from the keystore folder
	lastMod time.Time  // Last time instance when a file was modified
	mu      sync.RWMutex
}

// scan performs a new scan on the given directory, compares against the already
// cached filenames, and returns file sets: creates, deletes, updates.
func (fc *fileCache) scan(keyDir string) (mapset.Set, mapset.Set, mapset.Set, error) {
	t0 := time.Now()

	// List all the failes from the keystore folder
	files, err := os.ReadDir(keyDir)
	if err != nil {
		return nil, nil, nil, err
	}
	t1 := time.Now()

	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Iterate all the files and gather their metadata
	all := mapset.NewThreadUnsafeSet()
	mods := mapset.NewThreadUnsafeSet()

	var newLastMod time.Time
	for _, fi := range files {
		path := filepath.Join(keyDir, fi.Name())
		// Skip any non-key files from the folder
		if nonKeyFile(fi) {
			utils.Logger().Debug().Str("path", path).Msg("Ignoring file on account scan")
			continue
		}
		// Gather the set of all and freshly modified files
		all.Add(path)

		if info, err := fi.Info(); err != nil {
			continue
		} else {
			modified := info.ModTime()
			if modified.After(fc.lastMod) {
				mods.Add(path)
			}
			if modified.After(newLastMod) {
				newLastMod = modified
			}
		}
	}
	t2 := time.Now()

	// Update the tracked files and return the three sets
	deletes := fc.all.Difference(all)   // Deletes = previous - current
	creates := all.Difference(fc.all)   // Creates = current - previous
	updates := mods.Difference(creates) // Updates = modified - creates

	fc.all, fc.lastMod = all, newLastMod
	t3 := time.Now()

	// Report on the scanning stats and return
	utils.Logger().Debug().
		Uint64("list", uint64(t1.Sub(t0))).
		Uint64("set", uint64(t2.Sub(t1))).
		Uint64("diff", uint64(t3.Sub(t2))).
		Msg("FS scan times")
	return creates, deletes, updates, nil
}

// nonKeyFile ignores editor backups, hidden files and folders/symlinks.
func nonKeyFile(fi fs.DirEntry) bool {
	// Skip editor backups and UNIX-style hidden files.
	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
		return true
	}
	// Skip misc special files, directories (yes, symlinks too).
	if info, err := fi.Info(); err != nil {
		return true
	} else {
		if fi.IsDir() || info.Mode()&os.ModeType != 0 {
			return true
		}
	}
	return false
}
