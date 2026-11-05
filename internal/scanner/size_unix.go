//go:build unix

package scanner

import (
	"io/fs"
	"syscall"
)

type fileID struct{ dev, ino uint64 }

// diskUsage returns the bytes actually allocated to a file, its identity,
// and whether it has other hard links (so the caller can count it once).
func diskUsage(info fs.FileInfo) (size int64, id fileID, linked bool) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return info.Size(), fileID{}, false
	}
	// st_blocks is always in 512-byte units, regardless of fs block size.
	return int64(st.Blocks) * 512, fileID{uint64(st.Dev), uint64(st.Ino)}, st.Nlink > 1
}
