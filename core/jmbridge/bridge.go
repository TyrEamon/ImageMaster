package jmbridge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"ImageMaster/core/logger"
	"ImageMaster/core/types"
)

const (
	runtimeEnvVar = "IMAGEMASTER_JM_RUNTIME"
	pythonEnvVar  = "IMAGEMASTER_JM_PYTHON"
)

type bridgeCommand struct {
	Executable string
	Args       []string
	Source     string
}

type bridgeEvent struct {
	Type     string            `json:"type"`
	Name     string            `json:"name,omitempty"`
	Message  string            `json:"message,omitempty"`
	Phase    string            `json:"phase,omitempty"`
	Current  int               `json:"current,omitempty"`
	Total    int               `json:"total,omitempty"`
	SavePath string            `json:"savePath,omitempty"`
	Headers  map[string]string `json:"headers,omitempty"`
	Payload  json.RawMessage   `json:"payload,omitempty"`
}

type bridgeRunResult struct {
	payload     json.RawMessage
	savePath    string
	reportedErr string
	stderr      string
}

func Supports(rawURL string) bool {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	return strings.Contains(host, "18comic.vip") || strings.Contains(host, "18comic.org")
}

func HelperAvailable() bool {
	_, err := resolveBridgeCommand()
	return err == nil
}

func Download(ctx context.Context, updater types.TaskUpdater, rawURL, outputDir, proxy string) (string, error) {
	args := []string{
		"--target", strings.TrimSpace(rawURL),
		"--output", outputDir,
	}
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	result, err := runBridge(ctx, "download", args, updater)
	if err != nil {
		return "", err
	}

	if result.savePath == "" {
		return outputDir, nil
	}
	return result.savePath, nil
}

func Search(ctx context.Context, query, proxy string, page int) (SearchResult, error) {
	args := []string{
		"--query", strings.TrimSpace(query),
		"--page", fmt.Sprintf("%d", max(page, 1)),
	}
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	result, err := runBridge(ctx, "search", args, nil)
	if err != nil {
		return SearchResult{}, err
	}

	var payload SearchResult
	if err := json.Unmarshal(result.payload, &payload); err != nil {
		return SearchResult{}, fmt.Errorf("failed to decode JM search result: %w", err)
	}
	return payload, nil
}

func Ranking(ctx context.Context, kind, proxy string, page int) (RankingResult, error) {
	args := []string{
		"--kind", strings.TrimSpace(kind),
		"--page", fmt.Sprintf("%d", max(page, 1)),
	}
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	result, err := runBridge(ctx, "ranking", args, nil)
	if err != nil {
		return RankingResult{}, err
	}

	var payload RankingResult
	if err := json.Unmarshal(result.payload, &payload); err != nil {
		return RankingResult{}, fmt.Errorf("failed to decode JM ranking result: %w", err)
	}
	return payload, nil
}

func Detail(ctx context.Context, target, proxy string) (DetailResult, error) {
	args := []string{
		"--target", strings.TrimSpace(target),
	}
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	result, err := runBridge(ctx, "detail", args, nil)
	if err != nil {
		return DetailResult{}, err
	}

	var payload DetailResult
	if err := json.Unmarshal(result.payload, &payload); err != nil {
		return DetailResult{}, fmt.Errorf("failed to decode JM detail result: %w", err)
	}
	return payload, nil
}

func Images(ctx context.Context, target, proxy string) (ImageResult, error) {
	args := []string{
		"--target", strings.TrimSpace(target),
	}
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	result, err := runBridge(ctx, "images", args, nil)
	if err != nil {
		return ImageResult{}, err
	}

	var payload ImageResult
	if err := json.Unmarshal(result.payload, &payload); err != nil {
		return ImageResult{}, fmt.Errorf("failed to decode JM image result: %w", err)
	}
	return payload, nil
}

func ReadableImages(ctx context.Context, target, proxy, cacheDir string, retentionHours, sizeLimitMB int) (ImageResult, error) {
	args := []string{
		"--target", strings.TrimSpace(target),
	}
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}
	if cacheDir = strings.TrimSpace(cacheDir); cacheDir != "" {
		args = append(args, "--cache-dir", cacheDir)
	}
	if retentionHours > 0 {
		args = append(args, "--cache-retention-hours", fmt.Sprintf("%d", retentionHours))
	}
	if sizeLimitMB > 0 {
		args = append(args, "--cache-limit-mb", fmt.Sprintf("%d", sizeLimitMB))
	}

	result, err := runBridge(ctx, "read-images", args, nil)
	if err != nil {
		return ImageResult{}, err
	}

	var payload ImageResult
	if err := json.Unmarshal(result.payload, &payload); err != nil {
		return ImageResult{}, fmt.Errorf("failed to decode JM readable image result: %w", err)
	}
	return payload, nil
}

func runBridge(ctx context.Context, action string, extraArgs []string, updater types.TaskUpdater) (bridgeRunResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	cmdSpec, err := resolveBridgeCommand()
	if err != nil {
		return bridgeRunResult{}, err
	}

	args := append([]string{}, cmdSpec.Args...)
	args = append(args, "--action", action)
	args = append(args, extraArgs...)

	cmd := exec.CommandContext(ctx, cmdSpec.Executable, args...)
	hideConsoleWindow(cmd)
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUTF8=1",
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return bridgeRunResult{}, fmt.Errorf("failed to read JM helper stdout: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return bridgeRunResult{}, fmt.Errorf("failed to start JM helper (%s): %w", cmdSpec.Source, err)
	}

	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	result := bridgeRunResult{}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var event bridgeEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			logger.Info("JM helper output: %s", line)
			continue
		}

		switch event.Type {
		case "name":
			if updater != nil && event.Name != "" {
				updater.UpdateTaskName(event.Name)
			}
		case "status":
			if updater != nil {
				if event.Name != "" {
					updater.UpdateTaskName(event.Name)
				} else if event.Message != "" {
					updater.UpdateTaskName("JM - " + event.Message)
				}
				switch strings.ToLower(event.Phase) {
				case "download", "downloading":
					updater.UpdateTaskStatus(string(types.StatusDownloading), "")
				default:
					updater.UpdateTaskStatus(string(types.StatusParsing), "")
				}
			}
		case "progress":
			if updater != nil {
				updater.UpdateTaskStatus(string(types.StatusDownloading), "")
				updater.UpdateTaskProgress(event.Current, event.Total)
			}
		case "result":
			if len(event.Payload) > 0 {
				result.payload = event.Payload
			}
			if event.SavePath != "" {
				result.savePath = event.SavePath
				if updater != nil {
					updater.UpdateTaskField("savePath", result.savePath)
				}
			}
			if event.Name != "" && updater != nil {
				updater.UpdateTaskName(event.Name)
			}
		case "error":
			if event.Message != "" {
				result.reportedErr = event.Message
			}
		default:
			logger.Debug("ignored JM helper event: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return bridgeRunResult{}, fmt.Errorf("failed to read JM helper output: %w", err)
	}

	result.stderr = strings.TrimSpace(stderr.String())
	if err := cmd.Wait(); err != nil {
		message := strings.TrimSpace(result.reportedErr)
		if message == "" {
			message = result.stderr
		}
		if message == "" {
			message = err.Error()
		}
		return bridgeRunResult{}, fmt.Errorf("JM helper failed: %s", message)
	}

	return result, nil
}

func resolveBridgeCommand() (*bridgeCommand, error) {
	if path := strings.TrimSpace(os.Getenv(runtimeEnvVar)); path != "" {
		if isFile(path) {
			return &bridgeCommand{
				Executable: path,
				Source:     "env runtime",
			}, nil
		}
		return nil, fmt.Errorf("%s points to a missing file: %s", runtimeEnvVar, path)
	}

	for _, baseDir := range candidateBaseDirs() {
		for _, name := range []string{"imagemaster-jm-runtime.exe", "jm_bridge.exe"} {
			path := filepath.Join(baseDir, "runtime", name)
			if isFile(path) {
				return &bridgeCommand{
					Executable: path,
					Source:     filepath.Base(path),
				}, nil
			}
		}

		scriptPath := filepath.Join(baseDir, "runtime", "jm_bridge.py")
		if !isFile(scriptPath) {
			continue
		}

		pythonExe, pythonArgs, err := resolvePythonCommand()
		if err != nil {
			continue
		}
		if !pythonCanImportJM(pythonExe, pythonArgs) {
			continue
		}

		return &bridgeCommand{
			Executable: pythonExe,
			Args:       append(pythonArgs, scriptPath),
			Source:     "python script",
		}, nil
	}

	return nil, errors.New("JM helper not found")
}

func resolvePythonCommand() (string, []string, error) {
	if configured := strings.TrimSpace(os.Getenv(pythonEnvVar)); configured != "" {
		if isFile(configured) || hasCommand(configured) {
			return configured, nil, nil
		}
	}

	if path, err := exec.LookPath("python"); err == nil {
		return path, nil, nil
	}

	if path, err := exec.LookPath("py"); err == nil {
		return path, []string{"-3"}, nil
	}

	return "", nil, errors.New("python not found")
}

func pythonCanImportJM(executable string, baseArgs []string) bool {
	args := append([]string{}, baseArgs...)
	args = append(args, "-c", "import jmcomic")
	cmd := exec.Command(executable, args...)
	hideConsoleWindow(cmd)
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUTF8=1",
	)
	return cmd.Run() == nil
}

func candidateBaseDirs() []string {
	seen := make(map[string]struct{})
	dirs := make([]string, 0, 4)

	add := func(path string) {
		path = strings.TrimSpace(path)
		if path == "" {
			return
		}
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			return
		}
		seen[clean] = struct{}{}
		dirs = append(dirs, clean)
	}

	if exePath, err := os.Executable(); err == nil {
		add(filepath.Dir(exePath))
	}
	if cwd, err := os.Getwd(); err == nil {
		add(cwd)
	}

	return dirs
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
