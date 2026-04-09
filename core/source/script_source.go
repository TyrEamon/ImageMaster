package source

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
)

type ScriptSource struct {
	manifest   SourceManifest
	scriptPath string
	scriptBody string
	client     *http.Client
	summary    Summary
}

func NewScriptSource(manifest SourceManifest) (*ScriptSource, error) {
	scriptPath, err := resolveScriptPath(manifest)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("read script source %s: %w", scriptPath, err)
	}

	return &ScriptSource{
		manifest:   manifest,
		scriptPath: scriptPath,
		scriptBody: string(data),
		client:     &http.Client{Timeout: 20 * time.Second},
		summary: Summary{
			ID:           manifest.ID,
			Name:         manifest.Name,
			Type:         manifest.Type,
			Language:     manifest.Language,
			Website:      manifest.Website,
			Version:      manifest.Version,
			BuiltIn:      false,
			Capabilities: append([]string{}, manifest.Capabilities...),
			RankingKinds: append([]string{}, manifest.RankingKinds...),
			Description:  manifest.Description,
		},
	}, nil
}

func (s *ScriptSource) Summary() Summary {
	return s.summary
}

func (s *ScriptSource) Search(query string, page int) (SearchResult, error) {
	exported, err := s.call("search", query, page)
	if err != nil {
		return SearchResult{}, err
	}

	var result SearchResult
	if err := decodeScriptResult(exported, &result); err != nil {
		return SearchResult{}, err
	}
	result.Source = s.summary
	return result, nil
}

func (s *ScriptSource) Detail(itemID string) (DetailResult, error) {
	exported, err := s.call("detail", itemID)
	if err != nil {
		return DetailResult{}, err
	}

	var result DetailResult
	if err := decodeScriptResult(exported, &result); err != nil {
		return DetailResult{}, err
	}
	result.Source = s.summary
	return result, nil
}

func (s *ScriptSource) Images(chapterID string) (ImageResult, error) {
	exported, err := s.call("images", chapterID)
	if err != nil {
		return ImageResult{}, err
	}

	var result ImageResult
	if err := decodeScriptResult(exported, &result); err != nil {
		return ImageResult{}, err
	}
	result.Source = s.summary
	return result, nil
}

func (s *ScriptSource) Ranking(kind string, page int) (RankingResult, error) {
	exported, err := s.call("ranking", kind, page)
	if err != nil {
		return RankingResult{}, err
	}

	var result RankingResult
	if err := decodeScriptResult(exported, &result); err != nil {
		return RankingResult{}, err
	}
	result.Source = s.summary
	return result, nil
}

func (s *ScriptSource) call(method string, args ...any) (any, error) {
	vm := otto.New()
	ctxObject, err := s.buildContext(vm)
	if err != nil {
		return nil, err
	}

	if err := vm.Set("ctx", ctxObject); err != nil {
		return nil, fmt.Errorf("init script context: %w", err)
	}

	if _, err := vm.Run(s.scriptBody); err != nil {
		return nil, fmt.Errorf("run source script %s: %w", s.scriptPath, err)
	}

	sourceValue, err := vm.Get("source")
	if err != nil {
		return nil, fmt.Errorf("source script %s did not expose global source object: %w", s.scriptPath, err)
	}
	if !sourceValue.IsObject() {
		return nil, fmt.Errorf("source script %s did not expose a valid source object", s.scriptPath)
	}

	sourceObject := sourceValue.Object()
	if sourceObject == nil {
		return nil, fmt.Errorf("source script %s did not expose a valid source object", s.scriptPath)
	}

	if fn, err := sourceObject.Get(method); err != nil || !fn.IsFunction() {
		return nil, fmt.Errorf("source %s does not implement %s", s.summary.ID, method)
	}

	callArgs := make([]any, 0, len(args)+1)
	callArgs = append(callArgs, args...)
	callArgs = append(callArgs, ctxObject)

	resultValue, err := sourceObject.Call(method, callArgs...)
	if err != nil {
		return nil, fmt.Errorf("call source %s.%s: %w", s.summary.ID, method, err)
	}

	exported, err := resultValue.Export()
	if err != nil {
		return nil, fmt.Errorf("export source %s.%s result: %w", s.summary.ID, method, err)
	}

	return exported, nil
}

func (s *ScriptSource) buildContext(vm *otto.Otto) (*otto.Object, error) {
	ctxObject, err := vm.Object(`({})`)
	if err != nil {
		return nil, err
	}

	_ = ctxObject.Set("getJSON", func(call otto.FunctionCall) otto.Value {
		result, callErr := s.requestJSON(call.Argument(0).String(), exportHeaderMap(call.Argument(1)))
		if callErr != nil {
			panic(vm.MakeCustomError("SourceRequestError", callErr.Error()))
		}

		value, valueErr := vm.ToValue(result)
		if valueErr != nil {
			panic(vm.MakeCustomError("SourceRuntimeError", valueErr.Error()))
		}
		return value
	})

	_ = ctxObject.Set("getText", func(call otto.FunctionCall) otto.Value {
		result, callErr := s.requestText(call.Argument(0).String(), exportHeaderMap(call.Argument(1)))
		if callErr != nil {
			panic(vm.MakeCustomError("SourceRequestError", callErr.Error()))
		}

		value, valueErr := vm.ToValue(result)
		if valueErr != nil {
			panic(vm.MakeCustomError("SourceRuntimeError", valueErr.Error()))
		}
		return value
	})

	_ = ctxObject.Set("postJSON", func(call otto.FunctionCall) otto.Value {
		result, callErr := s.requestJSONWithBody(
			http.MethodPost,
			call.Argument(0).String(),
			exportBodyValue(call.Argument(1)),
			exportHeaderMap(call.Argument(2)),
		)
		if callErr != nil {
			panic(vm.MakeCustomError("SourceRequestError", callErr.Error()))
		}

		value, valueErr := vm.ToValue(result)
		if valueErr != nil {
			panic(vm.MakeCustomError("SourceRuntimeError", valueErr.Error()))
		}
		return value
	})

	_ = ctxObject.Set("urlQuery", func(call otto.FunctionCall) otto.Value {
		value, _ := vm.ToValue(url.QueryEscape(call.Argument(0).String()))
		return value
	})

	_ = ctxObject.Set("resolveURL", func(call otto.FunctionCall) otto.Value {
		resolved := resolveURL(call.Argument(0).String(), call.Argument(1).String())
		value, _ := vm.ToValue(resolved)
		return value
	})

	_ = ctxObject.Set("log", func(call otto.FunctionCall) otto.Value {
		return otto.UndefinedValue()
	})

	return ctxObject, nil
}

func (s *ScriptSource) requestJSON(rawURL string, headers map[string]string) (any, error) {
	text, err := s.requestText(rawURL, headers)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("decode json from %s: %w", rawURL, err)
	}

	return result, nil
}

func (s *ScriptSource) requestJSONWithBody(method string, rawURL string, body any, headers map[string]string) (any, error) {
	text, err := s.requestTextWithBody(method, rawURL, body, headers)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("decode json from %s: %w", rawURL, err)
	}

	return result, nil
}

func (s *ScriptSource) requestText(rawURL string, headers map[string]string) (string, error) {
	return s.requestTextWithBody(http.MethodGet, rawURL, nil, headers)
}

func (s *ScriptSource) requestTextWithBody(method string, rawURL string, body any, headers map[string]string) (string, error) {
	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("encode request body for %s: %w", rawURL, err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(strings.ToUpper(strings.TrimSpace(method)), strings.TrimSpace(rawURL), bodyReader)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "ImageMaster/0.2")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		req.Header.Set(trimmedKey, trimmedValue)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("request failed: %s %s", resp.Status, strings.TrimSpace(string(body)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func resolveScriptPath(manifest SourceManifest) (string, error) {
	script := strings.TrimSpace(manifest.Script)
	if script == "" {
		return "", fmt.Errorf("source %s has no local script configured", manifest.ID)
	}

	if filepath.IsAbs(script) {
		return filepath.Clean(script), nil
	}

	if manifest.ManifestPath == "" {
		return "", fmt.Errorf("source %s has no manifest path for resolving script", manifest.ID)
	}

	baseDir := filepath.Dir(manifest.ManifestPath)
	return filepath.Clean(filepath.Join(baseDir, filepath.FromSlash(script))), nil
}

func exportHeaderMap(value otto.Value) map[string]string {
	if !value.IsDefined() || value.IsNull() {
		return nil
	}

	exported, err := value.Export()
	if err != nil {
		return nil
	}

	if exportedMap, ok := exported.(map[string]any); ok {
		headers := make(map[string]string, len(exportedMap))
		for key, item := range exportedMap {
			headers[key] = strings.TrimSpace(fmt.Sprint(item))
		}
		return headers
	}

	return nil
}

func exportBodyValue(value otto.Value) any {
	if !value.IsDefined() || value.IsNull() {
		return nil
	}

	exported, err := value.Export()
	if err != nil {
		return nil
	}

	return exported
}

func resolveURL(baseRaw string, targetRaw string) string {
	baseRaw = strings.TrimSpace(baseRaw)
	targetRaw = strings.TrimSpace(targetRaw)
	if targetRaw == "" {
		return baseRaw
	}
	if baseRaw == "" {
		return targetRaw
	}

	baseURL, err := url.Parse(baseRaw)
	if err != nil {
		return targetRaw
	}
	targetURL, err := url.Parse(targetRaw)
	if err != nil {
		return targetRaw
	}

	return baseURL.ResolveReference(targetURL).String()
}

func decodeScriptResult[T any](exported any, target *T) error {
	data, err := json.Marshal(exported)
	if err != nil {
		return fmt.Errorf("marshal script result: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("decode script result: %w", err)
	}

	return nil
}
