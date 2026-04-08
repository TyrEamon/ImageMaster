package library

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"ImageMaster/core/types"
)

// Manga 漫画信息结构
type Manga struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	PreviewImg  string   `json:"previewImg"`
	ImagesCount int      `json:"imagesCount"`
	Images      []string `json:"images,omitempty"`
}

// MediaInfo 媒体信息
type MediaInfo struct {
	URL      string
	Filename string
}

// Manager 图书馆管理器
type Manager struct {
	ctx           context.Context
	configManager types.ConfigManager
	mediaTypes    map[string]bool
	outputDir     string
	mangas        []Manga
}

// NewManager 创建新的图书馆管理器
func NewManager(configManager types.ConfigManager, outputDir string) *Manager {
	// 默认支持的图片格式
	validExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true,
		".gif": true, ".webp": true, ".bmp": true,
	}

	return &Manager{
		configManager: configManager,
		mediaTypes:    validExts,
		outputDir:     outputDir,
		mangas:        []Manga{},
	}
}

// SetContext 设置上下文
func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

// GetLibraries 获取所有图书馆路径
func (m *Manager) GetLibraries() []string {
	return m.configManager.GetLibraries()
}

// LoadAllLibraries 加载所有图书馆
func (m *Manager) LoadAllLibraries() {
	m.mangas = []Manga{}
	for _, lib := range m.configManager.GetLibraries() {
		m.LoadLibrary(lib)
	}
}

// LoadLibrary 加载指定图书馆
func (m *Manager) LoadLibrary(path string) bool {
	m.mangas = []Manga{}
	return m.LoadMangaLibrary(path, &m.mangas)
}

// GetAllMangas 获取所有漫画
func (m *Manager) GetAllMangas() []Manga {
	return m.mangas
}

// DeleteManga 删除漫画（删除文件夹）
func (m *Manager) DeleteManga(path string) bool {
	err := os.RemoveAll(path)
	if err != nil {
		return false
	}

	// 从manga列表中移除
	for i, manga := range m.mangas {
		if manga.Path == path {
			m.mangas = append(m.mangas[:i], m.mangas[i+1:]...)
			break
		}
	}

	return true
}

// SetOutputDir 设置输出目录
func (m *Manager) SetOutputDir(dir string) {
	m.outputDir = dir
}

// GetOutputDir 获取输出目录
func (m *Manager) GetOutputDir() string {
	return m.outputDir
}

// LoadMangaLibrary 加载漫画库
func (m *Manager) LoadMangaLibrary(rootPath string, mangas *[]Manga) bool {
	// 递归获取文件夹下的所有子文件夹
	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 只处理文件夹
		if !d.IsDir() {
			return nil
		}

		// 跳过根路径
		if path == rootPath {
			return nil
		}

		// 获取文件夹中的图片
		images, err := m.GetImagesInDir(path)
		if err != nil || len(images) == 0 {
			return nil
		}

		// 排序图片
		m.SortImages(images)

		// 创建漫画信息
		manga := Manga{
			Name:        filepath.Base(path),
			Path:        path,
			PreviewImg:  images[0],
			ImagesCount: len(images),
			Images:      nil, // 不预加载所有图片路径
		}

		*mangas = append(*mangas, manga)

		return nil
	})

	return err == nil
}

// GetImagesInDir 获取指定目录中的所有图片
func (m *Manager) GetImagesInDir(dirPath string) ([]string, error) {
	var images []string

	// 读取目录中的所有文件
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if m.mediaTypes[ext] {
			images = append(images, filepath.Join(dirPath, entry.Name()))
		}
	}

	return images, nil
}

// GetMangaImages 获取指定漫画的所有图片
func (m *Manager) GetMangaImages(path string) []string {
	images, _ := m.GetImagesInDir(path)
	m.SortImages(images)
	return images
}

// SortImages 排序图片文件
func (m *Manager) SortImages(images []string) {
	sort.Slice(images, func(i, j int) bool {
		nameI := filepath.Base(images[i])
		nameJ := filepath.Base(images[j])

		// 尝试提取 page_offset 格式
		partsI := strings.Split(strings.TrimSuffix(nameI, filepath.Ext(nameI)), "_")
		partsJ := strings.Split(strings.TrimSuffix(nameJ, filepath.Ext(nameJ)), "_")

		if len(partsI) == 2 && len(partsJ) == 2 {
			pageI, errI1 := strconv.Atoi(partsI[0])
			offsetI, errI2 := strconv.Atoi(partsI[1])
			pageJ, errJ1 := strconv.Atoi(partsJ[0])
			offsetJ, errJ2 := strconv.Atoi(partsJ[1])

			if errI1 == nil && errI2 == nil && errJ1 == nil && errJ2 == nil {
				if pageI != pageJ {
					return pageI < pageJ
				}
				return offsetI < offsetJ
			}
		}

		// 回退到提取数字排序逻辑
		reNum := regexp.MustCompile(`\d+`)
		numsI := reNum.FindAllString(nameI, -1)
		numsJ := reNum.FindAllString(nameJ, -1)

		if len(numsI) > 0 && len(numsJ) > 0 {
			numI, _ := strconv.Atoi(numsI[0])
			numJ, _ := strconv.Atoi(numsJ[0])
			return numI < numJ
		}

		if len(numsI) > 0 {
			return true
		}
		if len(numsJ) > 0 {
			return false
		}

		return nameI < nameJ
	})
}

// GetImageDataUrl 获取图片的DataURL
func (m *Manager) GetImageDataUrl(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// 获取MIME类型
	ext := strings.ToLower(filepath.Ext(path))
	mimeType := "image/jpeg" // 默认
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".bmp":
		mimeType = "image/bmp"
	}

	// 构建data URL
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

// EnsureDir 确保目录存在
func (m *Manager) EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// IsImageFile 检查文件是否为图片
func (m *Manager) IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return m.mediaTypes[ext]
}
