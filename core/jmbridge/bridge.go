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
	cmdSpec, err := resolveBridgeCommand()
	if err != nil {
		return "", err
	}

	if updater != nil {
		updater.UpdateTaskName("JM - preparing helper")
		updater.UpdateTaskStatus(string(types.StatusParsing), "")
	}

	args := append([]string{}, cmdSpec.Args...)
	args = append(args,
		"--action", "download",
		"--target", strings.TrimSpace(rawURL),
		"--output", outputDir,
	)
	if proxy = strings.TrimSpace(proxy); proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	cmd := exec.CommandContext(ctx, cmdSpec.Executable, args...)
	cmd.Env = append(os.Environ(),
		"PYTHONIOENCODING=utf-8",
		"PYTHONUTF8=1",
	)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to read JM helper stdout: %w", err)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start JM helper (%s): %w", cmdSpec.Source, err)
	}

	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		savePath    string
		reportedErr string
	)

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
			if updater == nil {
				continue
			}
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
		case "progress":
			if updater != nil {
				updater.UpdateTaskStatus(string(types.StatusDownloading), "")
				updater.UpdateTaskProgress(event.Current, event.Total)
			}
		case "result":
			if event.Name != "" && updater != nil {
				updater.UpdateTaskName(event.Name)
			}
			if event.SavePath != "" {
				savePath = event.SavePath
				if updater != nil {
					updater.UpdateTaskField("savePath", savePath)
				}
			}
		case "error":
			if event.Message != "" {
				reportedErr = event.Message
			}
		default:
			logger.Debug("ignored JM helper event: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed to read JM helper output: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		message := strings.TrimSpace(reportedErr)
		if message == "" {
			message = strings.TrimSpace(stderr.String())
		}
		if message == "" {
			message = err.Error()
		}
		return "", fmt.Errorf("JM helper failed: %s", message)
	}

	if savePath == "" {
		savePath = outputDir
	}

	return savePath, nil
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
