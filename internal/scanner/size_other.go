//go:build !unix

package scanner

import "io/fs"

type fileID struct{ dev, ino uint64 }

func diskUsage(info fs.FileInfo) (size int64, id fileID, linked bool) {
	return info.Size(), fileID{}, false
}
