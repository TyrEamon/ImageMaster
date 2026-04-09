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

type Manga struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	PreviewImg  string   `json:"previewImg"`
	ImagesCount int      `json:"imagesCount"`
	Images      []string `json:"images,omitempty"`
}

type MediaInfo struct {
	URL      string
	Filename string
}

type Manager struct {
	ctx           context.Context
	configManager types.ConfigManager
	mediaTypes    map[string]bool
	outputDir     string
	mangas        []Manga
}

func NewManager(configManager types.ConfigManager, outputDir string) *Manager {
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

func (m *Manager) SetContext(ctx context.Context) {
	m.ctx = ctx
}

func (m *Manager) GetLibraries() []string {
	return m.configManager.GetLibraries()
}

func (m *Manager) LoadAllLibraries() {
	m.mangas = []Manga{}
	for _, lib := range m.configManager.GetLibraries() {
		m.LoadLibrary(lib)
	}
}

func (m *Manager) LoadLibrary(path string) bool {
	m.mangas = []Manga{}
	return m.LoadMangaLibrary(path, &m.mangas)
}

func (m *Manager) GetAllMangas() []Manga {
	return m.mangas
}

func (m *Manager) DeleteManga(path string) bool {
	err := os.RemoveAll(path)
	if err != nil {
		return false
	}

	for i, manga := range m.mangas {
		if manga.Path == path {
			m.mangas = append(m.mangas[:i], m.mangas[i+1:]...)
			break
		}
	}

	return true
}

func (m *Manager) SetOutputDir(dir string) {
	m.outputDir = dir
}

func (m *Manager) GetOutputDir() string {
	return m.outputDir
}

func (m *Manager) LoadMangaLibrary(rootPath string, mangas *[]Manga) bool {
	claimedRoots := map[string]struct{}{}

	rootEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return false
	}

	sort.Slice(rootEntries, func(i, j int) bool {
		return rootEntries[i].Name() < rootEntries[j].Name()
	})

	for _, entry := range rootEntries {
		if !entry.IsDir() {
			continue
		}

		childPath := filepath.Join(rootPath, entry.Name())
		directImages, directErr := m.GetImagesInDir(childPath)
		if directErr == nil && len(directImages) > 0 {
			m.SortImages(directImages)
			m.appendManga(childPath, directImages, mangas)
			claimedRoots[childPath] = struct{}{}
			continue
		}

		groupedImages, groupedErr := m.GetImagesInImmediateChildDirs(childPath)
		if groupedErr == nil && len(groupedImages) > 0 {
			m.appendManga(childPath, groupedImages, mangas)
			claimedRoots[childPath] = struct{}{}
		}
	}

	walkErr := filepath.WalkDir(rootPath, func(currentPath string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}
		if currentPath == rootPath {
			return nil
		}
		if _, claimed := claimedRoots[currentPath]; claimed {
			return filepath.SkipDir
		}

		images, imagesErr := m.GetImagesInDir(currentPath)
		if imagesErr != nil || len(images) == 0 {
			return nil
		}

		m.SortImages(images)
		m.appendManga(currentPath, images, mangas)
		return filepath.SkipDir
	})

	return walkErr == nil
}

func (m *Manager) GetImagesInDir(dirPath string) ([]string, error) {
	var images []string

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

func (m *Manager) GetMangaImages(path string) []string {
	images, _ := m.GetImagesInDir(path)
	if len(images) > 0 {
		m.SortImages(images)
		return images
	}

	groupedImages, _ := m.GetImagesInImmediateChildDirs(path)
	if len(groupedImages) > 0 {
		return groupedImages
	}

	return images
}

func (m *Manager) SortImages(images []string) {
	sort.Slice(images, func(i, j int) bool {
		nameI := filepath.Base(images[i])
		nameJ := filepath.Base(images[j])

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

func (m *Manager) GetImageDataUrl(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	mimeType := "image/jpeg"
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	case ".bmp":
		mimeType = "image/bmp"
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

func (m *Manager) EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

func (m *Manager) IsImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return m.mediaTypes[ext]
}

func (m *Manager) appendManga(path string, images []string, mangas *[]Manga) {
	if len(images) == 0 {
		return
	}

	*mangas = append(*mangas, Manga{
		Name:        filepath.Base(path),
		Path:        path,
		PreviewImg:  images[0],
		ImagesCount: len(images),
		Images:      nil,
	})
}

func (m *Manager) GetImagesInImmediateChildDirs(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	images := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		childPath := filepath.Join(dirPath, entry.Name())
		childImages, childErr := m.GetImagesInDir(childPath)
		if childErr != nil || len(childImages) == 0 {
			continue
		}

		m.SortImages(childImages)
		images = append(images, childImages...)
	}

	return images, nil
}
