package utils

import (
	"os"
	"regexp"
	"strings"
)

// MkdirAll 创建目录，会自动规范化目录名称中的非法字符
// 保留磁盘符号（如 D:）和路径分隔符（\），但替换其他非法字符为下划线
func MkdirAll(dir string, perm os.FileMode) error {
	// 规范化目录路径
	normalizedDir := NormalizePath(dir)
	return os.MkdirAll(normalizedDir, perm)
}

// NormalizePath 规范化路径，替换非法字符为下划线
// 保留磁盘符号和路径分隔符
func NormalizePath(path string) string {
	if path == "" {
		return path
	}

	// 检查是否是Windows绝对路径（如 D:\abc）
	var drivePrefix string
	var remainingPath string
	var separator string
	var isAbsolute bool

	// 匹配Windows驱动器模式 (如 "D:", "C:", etc.)
	driveRegex := regexp.MustCompile(`^[A-Za-z]:[\\/]?`)
	if match := driveRegex.FindString(path); match != "" {
		drivePrefix = match
		remainingPath = path[len(match):]
		// 检查使用的分隔符类型
		if strings.Contains(match, "\\") {
			separator = "\\"
		} else {
			separator = "/"
		}
		isAbsolute = true
	} else {
		// 检查是否是Unix风格的绝对路径
		if strings.HasPrefix(path, "/") {
			isAbsolute = true
			remainingPath = path[1:] // 去掉开头的 /
			separator = "/"
		} else {
			remainingPath = path
			// 对于相对路径，检测使用的分隔符
			if strings.Contains(path, "\\") {
				separator = "\\"
			} else {
				separator = "/"
			}
		}
	}

	// 分割路径为各个部分
	parts := strings.Split(remainingPath, separator)

	// 规范化每个路径部分，同时过滤空字符串
	var normalizedParts []string
	for _, part := range parts {
		if part != "" {
			normalizedParts = append(normalizedParts, normalizePathPart(part))
		}
	}

	// 重新组合路径
	normalizedPath := strings.Join(normalizedParts, separator)

	// 添加前缀
	if drivePrefix != "" {
		return drivePrefix + normalizedPath
	} else if isAbsolute && drivePrefix == "" {
		// Unix风格绝对路径，添加开头的 /
		return separator + normalizedPath
	}

	return normalizedPath
}

// normalizePathPart 规范化单个路径部分，替换非法字符为下划线
func normalizePathPart(part string) string {
	// Windows文件名非法字符: < > : " | ? * 以及控制字符 (0-31)
	// 但我们需要保留路径分隔符，所以这里只处理文件名部分
	illegalChars := regexp.MustCompile(`[<>:"|?*\x00-\x1f]`)
	return illegalChars.ReplaceAllString(part, "_")
}
