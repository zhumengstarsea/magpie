//go:build !darwin

package dirent

func stats(string) (map[string]Stat, bool) { return nil, false }
