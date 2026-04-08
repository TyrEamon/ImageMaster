package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Windows drive with colon should be preserved",
			input:    "D:\\abc",
			expected: "D:\\abc",
		},
		{
			name:     "Windows drive with illegal chars in folder name",
			input:    "D:\\abc\\DDD",
			expected: "D:\\abc\\DDD",
		},
		{
			name:     "Illegal characters in folder name",
			input:    "D:\\abc<>:?*|\"test",
			expected: "D:\\abc_______test",
		},
		{
			name:     "Multiple folders with illegal chars",
			input:    "D:\\folder1<test\\folder2>test\\folder3:test",
			expected: "D:\\folder1_test\\folder2_test\\folder3_test",
		},
		{
			name:     "Relative path with illegal chars",
			input:    "folder<test\\subfolder>test",
			expected: "folder_test\\subfolder_test",
		},
		{
			name:     "Path with question mark and asterisk",
			input:    "C:\\downloads\\file?.txt*folder",
			expected: "C:\\downloads\\file_.txt_folder",
		},
		{
			name:     "Path with pipe character",
			input:    "D:\\test|folder\\sub|folder",
			expected: "D:\\test_folder\\sub_folder",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only drive letter",
			input:    "C:",
			expected: "C:",
		},
		{
			name:     "Drive with forward slash",
			input:    "D:/folder<test",
			expected: "D:/folder_test",
		},
		{
			name:     "Unix-style path with illegal chars",
			input:    "/home/user/folder<test>/subfolder>test",
			expected: "/home/user/folder_test_/subfolder_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePath(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizePathPart(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal folder name",
			input:    "normalfolder",
			expected: "normalfolder",
		},
		{
			name:     "Folder with less than symbol",
			input:    "folder<test",
			expected: "folder_test",
		},
		{
			name:     "Folder with greater than symbol",
			input:    "folder>test",
			expected: "folder_test",
		},
		{
			name:     "Folder with colon",
			input:    "folder:test",
			expected: "folder_test",
		},
		{
			name:     "Folder with double quote",
			input:    "folder\"test",
			expected: "folder_test",
		},
		{
			name:     "Folder with pipe",
			input:    "folder|test",
			expected: "folder_test",
		},
		{
			name:     "Folder with question mark",
			input:    "folder?test",
			expected: "folder_test",
		},
		{
			name:     "Folder with asterisk",
			input:    "folder*test",
			expected: "folder_test",
		},
		{
			name:     "Multiple illegal characters",
			input:    "folder<>:?*|\"test",
			expected: "folder_______test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePathPart(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePathPart(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMkdirAll(t *testing.T) {
	// 创建临时目录进行测试
	tempDir := t.TempDir()
	
	tests := []struct {
		name     string
		path     string
		expected string // 期望创建的实际路径
	}{
		{
			name:     "Normal path",
			path:     filepath.Join(tempDir, "normal", "folder"),
			expected: filepath.Join(tempDir, "normal", "folder"),
		},
		{
			name:     "Path with illegal characters",
			path:     filepath.Join(tempDir, "folder<test", "subfolder>test"),
			expected: filepath.Join(tempDir, "folder_test", "subfolder_test"),
		},
		{
			name:     "Path with multiple illegal characters",
			path:     filepath.Join(tempDir, "folder:?*test", "sub|folder"),
			expected: filepath.Join(tempDir, "folder___test", "sub_folder"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MkdirAll(tt.path, 0755)
			if err != nil {
				t.Errorf("MkdirAll(%q) returned error: %v", tt.path, err)
				return
			}

			// 检查实际创建的目录是否存在
			if _, err := os.Stat(tt.expected); os.IsNotExist(err) {
				t.Errorf("Expected directory %q was not created", tt.expected)
			}

			// 清理创建的目录
			os.RemoveAll(tt.expected)
		})
	}
}