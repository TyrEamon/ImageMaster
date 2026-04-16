package library

import (
	"os"
	"syscall"
	"unsafe"
)

func getPathCreatedAt(path string) (int64, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	var data syscall.Win32FileAttributeData
	if err := syscall.GetFileAttributesEx(pointer, syscall.GetFileExInfoStandard, (*byte)(unsafe.Pointer(&data))); err != nil {
		info, statErr := os.Stat(path)
		if statErr != nil {
			return 0, err
		}
		return info.ModTime().UnixMilli(), nil
	}

	return data.CreationTime.Nanoseconds() / 1e6, nil
}
