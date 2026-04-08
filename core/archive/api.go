package archive

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"ImageMaster/core/logger"
	"ImageMaster/core/types"
)

var archiveExtensions = map[string]bool{
	".zip": true,
	".cbz": true,
	".7z":  true,
	".rar": true,
}

var archiveStatusOrder = map[string]int{
	"pending":   0,
	"failed":    1,
	"extracted": 2,
}

type ArchiveItem struct {
	ArchivePath string `json:"archivePath"`
	ArchiveName string `json:"archiveName"`
	LibraryPath string `json:"libraryPath"`
	TargetDir   string `json:"targetDir"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
	SizeBytes   int64  `json:"sizeBytes"`
}

type ScanResult struct {
	Roots          []string      `json:"roots"`
	ActiveLibrary  string        `json:"activeLibrary"`
	BandizipPath   string        `json:"bandizipPath"`
	TotalCount     int           `json:"totalCount"`
	PendingCount   int           `json:"pendingCount"`
	ExtractedCount int           `json:"extractedCount"`
	FailedCount    int           `json:"failedCount"`
	Items          []ArchiveItem `json:"items"`
}

type ExtractActionResult struct {
	ArchivePath string `json:"archivePath"`
	ArchiveName string `json:"archiveName"`
	TargetDir   string `json:"targetDir"`
	Status      string `json:"status"`
	Message     string `json:"message"`
}

type ExtractSummary struct {
	BandizipPath   string                `json:"bandizipPath"`
	TotalCount     int                   `json:"totalCount"`
	ExtractedCount int                   `json:"extractedCount"`
	SkippedCount   int                   `json:"skippedCount"`
	FailedCount    int                   `json:"failedCount"`
	Results        []ExtractActionResult `json:"results"`
}

type API struct {
	configManager types.ConfigManager
}

func NewAPI(configManager types.ConfigManager) *API {
	return &API{
		configManager: configManager,
	}
}

func (a *API) ScanArchives() (*ScanResult, error) {
	roots := a.getScanRoots()
	result := &ScanResult{
		Roots:          roots,
		ActiveLibrary:  a.configManager.GetActiveLibrary(),
		BandizipPath:   a.detectBandizipPath(),
		Items:          []ArchiveItem{},
		PendingCount:   0,
		ExtractedCount: 0,
		FailedCount:    0,
	}

	for _, root := range roots {
		items, err := a.scanRoot(root)
		if err != nil {
			logger.Warn("Scan archive root failed: %s | %v", root, err)
			continue
		}
		result.Items = append(result.Items, items...)
	}

	sort.Slice(result.Items, func(i, j int) bool {
		orderI := archiveStatusOrder[result.Items[i].Status]
		orderJ := archiveStatusOrder[result.Items[j].Status]
		if orderI != orderJ {
			return orderI < orderJ
		}
		if result.Items[i].LibraryPath != result.Items[j].LibraryPath {
			return result.Items[i].LibraryPath < result.Items[j].LibraryPath
		}
		return result.Items[i].ArchivePath < result.Items[j].ArchivePath
	})

	result.TotalCount = len(result.Items)
	for _, item := range result.Items {
		switch item.Status {
		case "pending":
			result.PendingCount++
		case "extracted":
			result.ExtractedCount++
		default:
			result.FailedCount++
		}
	}

	return result, nil
}

func (a *API) ExtractPendingArchives() (*ExtractSummary, error) {
	scanResult, err := a.ScanArchives()
	if err != nil {
		return nil, err
	}

	bandizipPath := scanResult.BandizipPath
	if bandizipPath == "" {
		return nil, fmt.Errorf("未找到 Bandizip 控制台工具，请先在设置页配置 bz.exe 路径")
	}

	summary := &ExtractSummary{
		BandizipPath: bandizipPath,
		Results:      []ExtractActionResult{},
	}

	for _, item := range scanResult.Items {
		if item.Status != "pending" {
			continue
		}

		summary.TotalCount++
		result := a.extractArchiveItem(item, bandizipPath)
		summary.Results = append(summary.Results, result)

		switch result.Status {
		case "extracted":
			summary.ExtractedCount++
		case "skipped":
			summary.SkippedCount++
		default:
			summary.FailedCount++
		}
	}

	return summary, nil
}

func (a *API) ExtractArchive(archivePath string) (*ExtractActionResult, error) {
	item, err := a.findArchiveItem(archivePath)
	if err != nil {
		return nil, err
	}

	bandizipPath := a.detectBandizipPath()
	if bandizipPath == "" {
		return nil, fmt.Errorf("未找到 Bandizip 控制台工具，请先在设置页配置 bz.exe 路径")
	}

	result := a.extractArchiveItem(*item, bandizipPath)
	if result.Status == "failed" {
		return &result, fmt.Errorf(result.Message)
	}

	return &result, nil
}

func (a *API) DetectBandizipPath() string {
	return a.detectBandizipPath()
}

func (a *API) findArchiveItem(archivePath string) (*ArchiveItem, error) {
	scanResult, err := a.ScanArchives()
	if err != nil {
		return nil, err
	}

	for _, item := range scanResult.Items {
		if strings.EqualFold(item.ArchivePath, archivePath) {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("未找到压缩包: %s", archivePath)
}

func (a *API) scanRoot(root string) ([]ArchiveItem, error) {
	items := []ArchiveItem{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			logger.Warn("Walk archive path failed: %s | %v", path, walkErr)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if !archiveExtensions[ext] {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			items = append(items, ArchiveItem{
				ArchivePath: path,
				ArchiveName: d.Name(),
				LibraryPath: root,
				TargetDir:   resolveExtractDirectory(path, root),
				Status:      "failed",
				Reason:      fmt.Sprintf("读取压缩包信息失败: %v", err),
			})
			return nil
		}

		targetDir := resolveExtractDirectory(path, root)
		status, reason := describeArchiveState(targetDir, path)
		items = append(items, ArchiveItem{
			ArchivePath: path,
			ArchiveName: d.Name(),
			LibraryPath: root,
			TargetDir:   targetDir,
			Status:      status,
			Reason:      reason,
			SizeBytes:   info.Size(),
		})

		return nil
	})

	return items, err
}

func (a *API) extractArchiveItem(item ArchiveItem, bandizipPath string) ExtractActionResult {
	result := ExtractActionResult{
		ArchivePath: item.ArchivePath,
		ArchiveName: item.ArchiveName,
		TargetDir:   item.TargetDir,
	}

	status, _ := describeArchiveState(item.TargetDir, item.ArchivePath)
	if status == "extracted" {
		result.Status = "skipped"
		result.Message = "已检测到解压内容，已跳过"
		return result
	}

	if err := os.MkdirAll(item.TargetDir, 0o755); err != nil {
		result.Status = "failed"
		result.Message = fmt.Sprintf("创建解压目录失败: %v", err)
		return result
	}

	logger.Info("Extract archive: %s -> %s", item.ArchivePath, item.TargetDir)
	cmd := exec.Command(bandizipPath, "x", "-y", item.ArchivePath, "-o:"+item.TargetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Status = "failed"
		result.Message = formatCommandError(err, output)
		return result
	}

	status, reason := describeArchiveState(item.TargetDir, item.ArchivePath)
	if status != "extracted" {
		result.Status = "failed"
		result.Message = reason
		return result
	}

	result.Status = "extracted"
	result.Message = "解压完成"
	return result
}

func (a *API) getScanRoots() []string {
	roots := []string{}
	seen := map[string]bool{}

	addPath := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}

		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			absPath = path
		}

		if seen[absPath] {
			return
		}

		seen[absPath] = true
		roots = append(roots, absPath)
	}

	addPath(a.configManager.GetActiveLibrary())
	addPath(a.configManager.GetOutputDir())
	for _, library := range a.configManager.GetLibraries() {
		addPath(library)
	}

	return roots
}

func (a *API) detectBandizipPath() string {
	candidates := []string{}

	if configured := strings.TrimSpace(a.configManager.GetBandizipPath()); configured != "" {
		candidates = append(candidates, configured)
	}

	if programFiles := os.Getenv("ProgramFiles"); programFiles != "" {
		candidates = append(candidates,
			filepath.Join(programFiles, "Bandizip", "bz.exe"),
			filepath.Join(programFiles, "Bandisoft", "Bandizip", "bz.exe"),
		)
	}

	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		candidates = append(candidates, filepath.Join(localAppData, "Bandizip", "bz.exe"))
	}

	candidates = append(candidates,
		`D:\bandizip\bz.exe`,
		`C:\Program Files\Bandizip\bz.exe`,
		`C:\Program Files\Bandisoft\Bandizip\bz.exe`,
	)

	seen := map[string]bool{}
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" || seen[strings.ToLower(candidate)] {
			continue
		}
		seen[strings.ToLower(candidate)] = true

		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate
		}
	}

	return ""
}

func resolveExtractDirectory(archivePath string, rootPath string) string {
	parentPath := strings.TrimRight(filepath.Dir(archivePath), `\`)
	rootPath = strings.TrimRight(rootPath, `\`)

	if strings.EqualFold(parentPath, rootPath) {
		return filepath.Join(rootPath, strings.TrimSuffix(filepath.Base(archivePath), filepath.Ext(archivePath)))
	}

	return filepath.Dir(archivePath)
}

func describeArchiveState(targetDir string, archivePath string) (string, string) {
	extracted, err := hasExtractedContent(targetDir, archivePath)
	if err != nil {
		return "failed", fmt.Sprintf("检查解压状态失败: %v", err)
	}

	if extracted {
		return "extracted", "检测到子文件夹或非压缩包文件"
	}

	return "pending", "等待解压"
}

func hasExtractedContent(targetDir string, archivePath string) (bool, error) {
	info, err := os.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	if !info.IsDir() {
		return false, nil
	}

	children, err := os.ReadDir(targetDir)
	if err != nil {
		return false, err
	}

	for _, child := range children {
		childPath := filepath.Join(targetDir, child.Name())
		if child.IsDir() {
			return true, nil
		}

		if strings.EqualFold(childPath, archivePath) {
			continue
		}

		if !archiveExtensions[strings.ToLower(filepath.Ext(child.Name()))] {
			return true, nil
		}
	}

	return false, nil
}

func formatCommandError(err error, output []byte) string {
	message := strings.TrimSpace(string(output))
	if message == "" {
		return fmt.Sprintf("Bandizip 执行失败: %v", err)
	}

	return fmt.Sprintf("Bandizip 执行失败: %s", message)
}
