package dirent

import (
	"io/fs"
	"time"
)

// Stat is what a folder's listing says of one entry: its kind (Mode's type
// bits, 0 for a regular file), and a regular file's size and modification
// time, as os.Lstat gives them.
type Stat struct {
	Mode fs.FileMode
	Size int64
	Mod  time.Time
}

// Stats is each entry of dir by name, as os.Lstat gives it, read in a few
// calls for the whole folder where the system has them (getattrlistbulk on
// macOS), rather than one per entry. ok is false where the system has none
// or the folder couldn't be read, and the caller asks after each entry
// itself.
func Stats(dir string) (map[string]Stat, bool) {
	return stats(dir)
}
