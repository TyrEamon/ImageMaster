//go:build !windows

package library

import "os"

func getPathCreatedAt(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.ModTime().UnixMilli(), nil
}
