package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
)

const (
	maxChatVisionReviewImages = 3
	maxChatVisionImageBytes   = 4 * 1024 * 1024
)

type chatToolRuntime struct {
	CtxInfo chatContext
	Auth    chatAuthConfig
	Model   string
	Report  chatStatusSink
}

type openAIReviewRequest struct {
	Model string                `json:"model"`
	Input []openAIReviewMessage `json:"input"`
}

type openAIReviewMessage struct {
	Role    string              `json:"role"`
	Content []openAIReviewInput `json:"content"`
}

type openAIReviewInput struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

type chatVisionImage struct {
	AbsPath      string
	RelativePath string
	InstanceName string
	CameraName   string
}

type chatProcessProgressReporter struct {
	report          chatStatusSink
	currentInstance string
}

func (c *chatProcessProgressReporter) Update(msg string) {
	reportChatStatus(c.report, msg)
}

func (c *chatProcessProgressReporter) Done() {
	reportChatStatus(c.report, "OpenSCAD generation complete")
}

func (c *chatProcessProgressReporter) Error(err error) {
	if err == nil {
		return
	}
	reportChatStatus(c.report, "OpenSCAD generation failed: "+err.Error())
}

func (c *chatProcessProgressReporter) Construct(instances []models.InstanceConfig, nonSkippedInstances int) {
	if nonSkippedInstances == 0 {
		reportChatStatus(c.report, "No instances matched the current config")
		return
	}
	reportChatStatus(c.report, fmt.Sprintf("Queued %d instance(s) for generation", nonSkippedInstances))
}

func (c *chatProcessProgressReporter) StartInstance(instanceID string, name string, instanceIndex int, instanceCount int) {
	c.currentInstance = name
	reportChatStatus(c.report, fmt.Sprintf("Generating instance %s (%d/%d)", name, instanceIndex+1, instanceCount))
}

func (c *chatProcessProgressReporter) FinishInstance() {
	if strings.TrimSpace(c.currentInstance) == "" {
		return
	}
	reportChatStatus(c.report, "Finished instance "+c.currentInstance)
	c.currentInstance = ""
}

func buildProcessConfigTool() openAITool {
	stringSchema := map[string]any{"type": "string"}
	return openAITool{
		Type:        "function",
		Name:        "process_config_and_review",
		Description: "Validate and process a config.toml inside the current project root after editing it. Returns render errors, generated image paths, and an OpenAI vision review summary when images are available.",
		Strict:      true,
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path":          stringSchema,
				"review_images": map[string]any{"type": "boolean"},
				"max_images":    map[string]any{"type": "integer", "minimum": 1, "maximum": maxChatVisionReviewImages},
			},
			"required":             []string{"path"},
			"additionalProperties": false,
		},
	}
}

func processConfigAndReviewTool(ctx context.Context, runtime chatToolRuntime, requestedPath string, reviewImages bool, maxImages int) (map[string]any, error) {
	targetAbs, targetRel, err := resolveProjectPath(runtime.CtxInfo.ProjectRoot, requestedPath, false)
	if err != nil {
		return nil, err
	}
	if filepath.Ext(targetAbs) != ".toml" && filepath.Base(targetAbs) != "config.toml" {
		return nil, fmt.Errorf("path %q is not a config.toml file", requestedPath)
	}
	if maxImages <= 0 || maxImages > maxChatVisionReviewImages {
		maxImages = maxChatVisionReviewImages
	}

	reportChatStatus(runtime.Report, "Validating "+filepath.Base(targetRel))

	config, warn, loadErr := pkg.LoadConfigFromFile(models.CmdFlags{
		ConfigFile:   targetAbs,
		Server:       true,
		ServerFolder: strings.TrimSpace(runtime.CtxInfo.ServerFolder),
	})
	if loadErr != nil {
		return map[string]any{
			"ok":           true,
			"config_path":  filepath.ToSlash(targetRel),
			"config_valid": false,
			"warning":      warningString(warn),
			"load_error":   loadErr.Error(),
		}, nil
	}

	progress := &chatProcessProgressReporter{report: runtime.Report}
	reportChatStatus(runtime.Report, "Running OpenSCAD generation for "+filepath.Base(targetRel))

	result, processErr := pkg.Process(config, progress, nil, pkg.Operations{GenerateReport: false}, true)

	payload := map[string]any{
		"ok":                true,
		"config_path":       filepath.ToSlash(targetRel),
		"config_valid":      true,
		"warning":           warningString(warn),
		"process_succeeded": processErr == nil,
		"summary":           buildChatProcessSummary(result),
		"errors":            buildChatProcessErrors(result),
		"images":            buildChatProcessImageList(runtime.CtxInfo.ProjectRoot, result.ImageResults),
	}
	if processErr != nil {
		payload["process_error"] = processErr.Error()
	}

	if reviewImages {
		visionImages := selectVisionReviewImages(runtime.CtxInfo.ProjectRoot, result.ImageResults, maxImages)
		if len(visionImages) == 0 {
			payload["image_review"] = "No generated PNGs were available to review."
			return payload, nil
		}

		reportChatStatus(runtime.Report, fmt.Sprintf("Reviewing %d generated image(s)", len(visionImages)))
		review, reviewErr := reviewGeneratedImagesWithOpenAI(ctx, runtime, filepath.ToSlash(targetRel), visionImages)
		if reviewErr != nil {
			payload["image_review_error"] = reviewErr.Error()
		} else {
			payload["image_review"] = review
			payload["images_reviewed"] = len(visionImages)
		}
	}

	return payload, nil
}

func buildChatProcessSummary(result models.ProcessResult) map[string]any {
	totalInstances := len(result.Instances)
	completedInstances := 0
	successfulInstances := 0
	failedInstances := 0
	skippedInstances := 0

	for _, instance := range result.Instances {
		if instance.IsComplete {
			completedInstances++
		}
		if strings.TrimSpace(instance.SkippedReason) != "" {
			skippedInstances++
			continue
		}
		if instance.IsSuccessful && strings.TrimSpace(instance.ConfigError) == "" {
			successfulInstances++
			continue
		}
		if strings.TrimSpace(instance.ConfigError) != "" || (!instance.IsSuccessful && len(instance.STLResults) > 0) {
			failedInstances++
		}
	}

	return map[string]any{
		"config_file":          result.ConfigFile,
		"export_location":      result.ExportLocation,
		"total_instances":      totalInstances,
		"completed_instances":  completedInstances,
		"successful_instances": successfulInstances,
		"failed_instances":     failedInstances,
		"skipped_instances":    skippedInstances,
		"image_results":        len(result.ImageResults),
		"total_time_ms":        result.TotalTimeTaken.Milliseconds(),
	}
}

func buildChatProcessErrors(result models.ProcessResult) []map[string]any {
	errors := make([]map[string]any, 0)
	for _, instance := range result.Instances {
		if strings.TrimSpace(instance.SkippedReason) != "" {
			errors = append(errors, map[string]any{
				"instance": instance.AutoName,
				"type":     "skipped",
				"message":  instance.SkippedReason,
			})
			continue
		}

		if strings.TrimSpace(instance.ConfigError) != "" {
			errors = append(errors, map[string]any{
				"instance":               instance.AutoName,
				"type":                   "render_error",
				"message":                instance.ConfigError,
				"command_output_excerpt": buildSTLCommandOutputExcerpt(instance.STLResults),
			})
		}
	}
	return errors
}

func buildChatProcessImageList(projectRoot string, imageResults []models.GenerateImageResult) []map[string]any {
	out := make([]map[string]any, 0, len(imageResults))
	for _, imageResult := range imageResults {
		out = append(out, map[string]any{
			"instance":    imageResult.InstanceConfig.AutoName,
			"camera_name": imageResult.CameraName,
			"path":        filepath.ToSlash(makeProjectRelative(projectRoot, imageResult.OutputPath)),
		})
	}
	return out
}

func buildSTLCommandOutputExcerpt(results []models.GenerateSTLResult) string {
	for _, result := range results {
		if trimmed := strings.TrimSpace(result.OutputLog); trimmed != "" {
			excerpt, _ := trimToolOutput(trimmed, 2400)
			return excerpt
		}
		if trimmed := strings.TrimSpace(result.CommandOutput); trimmed != "" {
			excerpt, _ := trimToolOutput(trimmed, 2400)
			return excerpt
		}
	}
	return ""
}

func selectVisionReviewImages(projectRoot string, imageResults []models.GenerateImageResult, maxImages int) []chatVisionImage {
	candidates := make([]chatVisionImage, 0, len(imageResults))
	for _, imageResult := range imageResults {
		if strings.TrimSpace(imageResult.OutputPath) == "" {
			continue
		}
		if _, err := os.Stat(imageResult.OutputPath); err != nil {
			continue
		}
		candidates = append(candidates, chatVisionImage{
			AbsPath:      imageResult.OutputPath,
			RelativePath: filepath.ToSlash(makeProjectRelative(projectRoot, imageResult.OutputPath)),
			InstanceName: imageResult.InstanceConfig.AutoName,
			CameraName:   imageResult.CameraName,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		pi := chatVisionImagePriority(candidates[i])
		pj := chatVisionImagePriority(candidates[j])
		if pi != pj {
			return pi < pj
		}
		if candidates[i].InstanceName != candidates[j].InstanceName {
			return candidates[i].InstanceName < candidates[j].InstanceName
		}
		return candidates[i].RelativePath < candidates[j].RelativePath
	})

	selected := make([]chatVisionImage, 0, maxImages)
	totalBytes := int64(0)
	for _, candidate := range candidates {
		info, err := os.Stat(candidate.AbsPath)
		if err != nil {
			continue
		}
		if totalBytes+info.Size() > maxChatVisionImageBytes && len(selected) > 0 {
			break
		}
		selected = append(selected, candidate)
		totalBytes += info.Size()
		if len(selected) >= maxImages {
			break
		}
	}
	return selected
}

func chatVisionImagePriority(image chatVisionImage) int {
	name := strings.ToLower(strings.TrimSpace(image.CameraName))
	switch name {
	case "nice":
		return 0
	case "obj":
		return 1
	case "all":
		return 2
	default:
		return 10
	}
}

func reviewGeneratedImagesWithOpenAI(ctx context.Context, runtime chatToolRuntime, configPath string, images []chatVisionImage) (string, error) {
	content := []openAIReviewInput{
		{
			Type: "input_text",
			Text: "Review these generated OpenSCAD render images for visible geometry issues, obvious printability issues, and whether the latest config change appears to have helped. Keep it concise and concrete. Mention the relevant image paths when useful. Config path: " + configPath,
		},
	}

	for _, image := range images {
		dataURL, err := encodeImageDataURL(image.AbsPath)
		if err != nil {
			return "", err
		}
		content = append(content, openAIReviewInput{
			Type: "input_text",
			Text: fmt.Sprintf("Image: %s (%s)", image.RelativePath, image.CameraName),
		})
		content = append(content, openAIReviewInput{
			Type:     "input_image",
			ImageURL: dataURL,
			Detail:   "high",
		})
	}

	reqBody := openAIReviewRequest{
		Model: runtime.Model,
		Input: []openAIReviewMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("encode image review request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURLForProvider(runtime.Auth.SelectedProvider), "/")+"/v1/responses", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("build image review request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+runtime.Auth.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send image review request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read image review response: %w", err)
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("image review request failed: %s", formatOpenAIError(resp.StatusCode, body))
	}

	result, err := parseOpenAIChatResult(body)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(result.OutputText) == "" {
		return "", fmt.Errorf("image review response did not include text output")
	}
	return result.OutputText, nil
}

func encodeImageDataURL(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read image %s: %w", path, err)
	}
	return "data:" + mimeTypeForImagePath(path) + ";base64," + base64.StdEncoding.EncodeToString(body), nil
}

func mimeTypeForImagePath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func makeProjectRelative(projectRoot string, path string) string {
	if strings.TrimSpace(projectRoot) == "" {
		return path
	}
	rel, err := filepath.Rel(projectRoot, path)
	if err != nil {
		return path
	}
	return rel
}

func warningString(warn error) string {
	if warn == nil {
		return ""
	}
	return warn.Error()
}

func reportChatStatus(report chatStatusSink, message string) {
	if report == nil {
		return
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return
	}
	report(message)
}
