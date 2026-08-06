package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kiwikid/openscadgen/pkg"
	"github.com/kiwikid/openscadgen/pkg/models"
	"github.com/kiwikid/openscadgen/pkg/templates"
)

const (
	providerOpenAI         = "openai"
	providerDeepSeek       = "deepseek"
	defaultOpenAIModel     = "gpt-4.1"
	defaultDeepSeekModel   = "deepseek-v4-flash"
	defaultOpenAIBaseURL   = "https://api.openai.com"
	defaultDeepSeekBaseURL = "https://api.deepseek.com"
	maxChatHistoryMessages = 24
	maxChatContextBytes    = 16000
	maxChatToolRounds      = 12
	maxChatProjectEntries  = 40
	maxChatOutputBytes     = 24000
	maxChatReadBytes       = 20000
	maxChatDiffBytes       = 24000
)

const defaultChatInstructions = `You are assisting with OpenSCADGen projects.

Help the user with OpenSCAD model iteration, config.toml structure, parameter choices, naming, instance generation, and server-mode workflow.
Be concise, practical, and explicit about uncertainty.
If file context is attached, use it as primary grounding and say when you are inferring beyond it.`

type chatContext struct {
	ConfigEncoded       string
	ServerFolderEncoded string
	ConfigPath          string
	ServerFolder        string
	ResolvedPath        string
	ProjectRoot         string
	ProjectLabel        string
	ProjectSummary      string
	ContextLabel        string
	Content             string
	Truncated           bool
	LoadErr             error
	ProjectErr          error
}

type chatSubmission struct {
	History      []models.ChatMessage
	Draft        string
	Auth         chatAuthConfig
	Model        string
	CtxInfo      chatContext
	Instructions string
	Notice       string
}

type openAIResponsesRequest struct {
	Model             string            `json:"model"`
	Instructions      string            `json:"instructions,omitempty"`
	Input             []json.RawMessage `json:"input"`
	Tools             []openAITool      `json:"tools,omitempty"`
	ToolChoice        string            `json:"tool_choice,omitempty"`
	ParallelToolCalls bool              `json:"parallel_tool_calls,omitempty"`
}

type openAITool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Strict      bool           `json:"strict,omitempty"`
}

type openAIInputMessage struct {
	Role    string            `json:"role"`
	Content []openAITextInput `json:"content"`
}

type openAITextInput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIFunctionCallOutput struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

type openAIResponsesResponse struct {
	Error  *openAIResponseError `json:"error"`
	Output []openAIOutputItem   `json:"output"`
	Model  string               `json:"model"`
}

type deepSeekChatCompletionsRequest struct {
	Model    string                `json:"model"`
	Messages []deepSeekChatMessage `json:"messages"`
	Stream   bool                  `json:"stream"`
}

type deepSeekChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type deepSeekChatCompletionsResponse struct {
	Error   *openAIResponseError `json:"error"`
	Choices []deepSeekChoice     `json:"choices"`
	Model   string               `json:"model"`
}

type deepSeekChoice struct {
	Message deepSeekAssistantMessage `json:"message"`
}

type deepSeekAssistantMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIOutputItem struct {
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []openAIOutputText `json:"content"`
}

type openAIOutputText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type openAIErrorEnvelope struct {
	Error *openAIResponseError `json:"error"`
}

type openAIResponseError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type chatAuthConfig struct {
	SelectedProvider  string
	Token             string
	Label             string
	OpenAIAvailable   bool
	DeepSeekAvailable bool
}

type chatProviderResult struct {
	Provider    string            `json:"provider"`
	Model       string            `json:"model"`
	OutputText  string            `json:"output_text"`
	ToolCalls   []chatToolCall    `json:"tool_calls,omitempty"`
	RawResponse json.RawMessage   `json:"raw_response,omitempty"`
	OutputItems []json.RawMessage `json:"-"`
}

type chatToolCall struct {
	Type        string          `json:"type"`
	ID          string          `json:"id,omitempty"`
	CallID      string          `json:"call_id,omitempty"`
	Name        string          `json:"name,omitempty"`
	Arguments   string          `json:"arguments,omitempty"`
	Status      string          `json:"status,omitempty"`
	ServerLabel string          `json:"server_label,omitempty"`
	Raw         json.RawMessage `json:"raw,omitempty"`
}

type openAIResponsesRawEnvelope struct {
	Error  *openAIResponseError `json:"error"`
	Output []json.RawMessage    `json:"output"`
	Model  string               `json:"model"`
}

func handleChatRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleChatGet(w, r)
	case http.MethodPost:
		handleChatPost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleChatGet(w http.ResponseWriter, r *http.Request) {
	ctxInfo := buildChatContext(r, false)
	auth := getChatAuthConfig(readRequestedProvider(r))
	data := makeChatPageData(r, nil, auth, defaultModel(auth.SelectedProvider), "", "", "", ctxInfo)

	if !data.APIKeyConfigured {
		data.Notice = "Set DEEPSEEK_API_TOKEN or OPENAI_API_KEY to send prompts. You can still open this page without either."
	}

	renderChatPage(w, r, data)
}

func handleChatPost(w http.ResponseWriter, r *http.Request) {
	submission, historyErr, err := buildChatSubmission(r)
	if err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}
	if historyErr != nil {
		ctxInfo := buildChatContext(r, false)
		auth := getChatAuthConfig(readRequestedProvider(r))
		data := makeChatPageData(r, nil, auth, readChatModel(r, auth.SelectedProvider), submission.Draft, "Conversation state was invalid. Start a new chat.", "", ctxInfo)
		renderChatPage(w, r, data)
		return
	}
	if submission.Draft == "" {
		data := makeChatPageData(r, submission.History, submission.Auth, submission.Model, "", "Enter a message before sending.", submission.Notice, submission.CtxInfo)
		renderChatPage(w, r, data)
		return
	}

	history := append(append([]models.ChatMessage(nil), submission.History...), models.ChatMessage{
		Role:    "user",
		Content: submission.Draft,
	})
	history = normalizeChatHistory(history)
	if submission.Auth.Token == "" {
		data := makeChatPageData(r, history, submission.Auth, submission.Model, "", "Neither DEEPSEEK_API_TOKEN nor OPENAI_API_KEY is set.", submission.Notice, submission.CtxInfo)
		renderChatPage(w, r, data)
		return
	}

	answer, err := createProviderChatResponseWithContext(r.Context(), submission.Auth, submission.Model, submission.Instructions, history, submission.CtxInfo)
	if err != nil {
		log.Printf("chat response error: %v", err)
		data := makeChatPageData(r, history, submission.Auth, submission.Model, "", err.Error(), submission.Notice, submission.CtxInfo)
		renderChatPage(w, r, data)
		return
	}

	history = append(history, models.ChatMessage{
		Role:    "assistant",
		Content: answer,
	})
	history = normalizeChatHistory(history)

	data := makeChatPageData(r, history, submission.Auth, submission.Model, "", "", submission.Notice, submission.CtxInfo)
	renderChatPage(w, r, data)
}

func renderChatPage(w http.ResponseWriter, r *http.Request, data models.ChatPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var err error
	if r.Header.Get("HX-Request") == "true" {
		err = templates.ChatPanel(data).Render(r.Context(), w)
	} else {
		err = templates.ChatPage(data).Render(r.Context(), w)
	}
	if err != nil {
		http.Error(w, "Failed to render chat page: "+err.Error(), http.StatusInternalServerError)
	}
}

func makeChatPageData(r *http.Request, history []models.ChatMessage, auth chatAuthConfig, model string, draft string, errMsg string, notice string, ctxInfo chatContext) models.ChatPageData {
	history = normalizeChatHistory(history)
	historyJSON, err := json.Marshal(history)
	if err != nil {
		historyJSON = []byte("[]")
	}

	title := "AI Chat"
	if ctxInfo.ContextLabel != "" {
		title = fmt.Sprintf("AI Chat - %s", ctxInfo.ContextLabel)
	}
	contextFileLabel := ""
	if ctxInfo.ResolvedPath != "" {
		contextFileLabel = ctxInfo.ContextLabel
	}

	homeURL := pkg.BuildHomeURL(ctxInfo.ServerFolder)
	backURL := homeURL
	if ctxInfo.ConfigPath != "" {
		backURL = pkg.BuildConfigFileURL(ctxInfo.ConfigPath, ctxInfo.ServerFolder)
	}

	actionURL := buildChatRouteURL(ctxInfo.ConfigEncoded, ctxInfo.ServerFolderEncoded, auth.SelectedProvider)
	runURL := buildChatRunURL(ctxInfo.ConfigEncoded, ctxInfo.ServerFolderEncoded, auth.SelectedProvider)

	return models.ChatPageData{
		Title:                   title,
		Messages:                history,
		HistoryJSON:             string(historyJSON),
		Draft:                   draft,
		Provider:                auth.SelectedProvider,
		OpenAIAvailable:         auth.OpenAIAvailable,
		DeepSeekAvailable:       auth.DeepSeekAvailable,
		OpenAIDefaultModel:      defaultModel(providerOpenAI),
		DeepSeekDefaultModel:    defaultModel(providerDeepSeek),
		Model:                   model,
		Error:                   errMsg,
		Notice:                  notice,
		APIKeyConfigured:        auth.Token != "",
		APIKeyLabel:             auth.Label,
		ProjectLabel:            ctxInfo.ProjectLabel,
		ProjectPath:             ctxInfo.ProjectRoot,
		ProjectToolsEnabled:     chatProjectToolsEnabled(auth.SelectedProvider, ctxInfo),
		ContextFileLabel:        contextFileLabel,
		ContextFilePath:         ctxInfo.ResolvedPath,
		ContextIncluded:         ctxInfo.Content != "",
		AsyncEnabled:            true,
		ConfigAutomationEnabled: chatConfigAutomationEnabled(auth, ctxInfo),
		HomeURL:                 homeURL,
		BackURL:                 backURL,
		ActionURL:               actionURL,
		ResetURL:                actionURL,
		RunURL:                  runURL,
		StatusURL:               "/api/chat/status",
	}
}

func getChatAuthConfig(requestedProvider string) chatAuthConfig {
	deepseekToken := strings.TrimSpace(os.Getenv("DEEPSEEK_API_TOKEN"))
	openAIToken := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	openAIAvailable := openAIToken != ""
	deepSeekAvailable := deepseekToken != ""
	selectedProvider := resolveSelectedProvider(requestedProvider, openAIAvailable, deepSeekAvailable)

	switch selectedProvider {
	case providerDeepSeek:
		return chatAuthConfig{
			SelectedProvider:  selectedProvider,
			Token:             deepseekToken,
			Label:             selectedProviderLabel(selectedProvider, deepSeekAvailable),
			OpenAIAvailable:   openAIAvailable,
			DeepSeekAvailable: deepSeekAvailable,
		}
	case providerOpenAI:
		return chatAuthConfig{
			SelectedProvider:  selectedProvider,
			Token:             openAIToken,
			Label:             selectedProviderLabel(selectedProvider, openAIAvailable),
			OpenAIAvailable:   openAIAvailable,
			DeepSeekAvailable: deepSeekAvailable,
		}
	default:
		return chatAuthConfig{
			SelectedProvider:  selectedProvider,
			Label:             "No API token configured",
			OpenAIAvailable:   openAIAvailable,
			DeepSeekAvailable: deepSeekAvailable,
		}
	}
}

func buildChatContext(r *http.Request, loadContent bool) chatContext {
	query := r.URL.Query()
	configEncoded := strings.TrimSpace(query.Get("config"))
	serverEncoded := strings.TrimSpace(query.Get("server_folder"))

	ctxInfo := buildChatContextFromPaths(decodeMaybeBase64(configEncoded), decodeMaybeBase64(serverEncoded), loadContent)
	ctxInfo.ConfigEncoded = configEncoded
	ctxInfo.ServerFolderEncoded = serverEncoded
	return ctxInfo
}

func buildChatContextFromPaths(configPath string, serverFolder string, loadContent bool) chatContext {
	configPath = strings.TrimSpace(configPath)
	serverFolder = strings.TrimSpace(serverFolder)

	resolvedPath := ""
	if configPath != "" {
		resolvedPath = resolveConfigPath(configPath, serverFolder)
	}

	projectRoot := resolveChatProjectRoot(resolvedPath, serverFolder)
	projectLabel := filepath.Base(projectRoot)
	if projectLabel == "" || projectLabel == "." || projectLabel == string(filepath.Separator) {
		projectLabel = projectRoot
	}

	contextLabel := ""
	switch {
	case resolvedPath != "":
		contextLabel = filepath.Base(resolvedPath)
		if contextLabel == "" || contextLabel == "." || contextLabel == string(filepath.Separator) {
			contextLabel = configPath
		}
	case projectLabel != "":
		contextLabel = projectLabel
	}

	out := chatContext{
		ConfigPath:   configPath,
		ServerFolder: serverFolder,
		ResolvedPath: resolvedPath,
		ProjectRoot:  projectRoot,
		ProjectLabel: projectLabel,
		ContextLabel: contextLabel,
	}

	if !loadContent {
		return out
	}

	if resolvedPath != "" {
		content, truncated, err := loadContextFile(resolvedPath, maxChatContextBytes)
		out.Content = content
		out.Truncated = truncated
		out.LoadErr = err
	}
	if projectRoot != "" {
		projectSummary, err := loadProjectSummary(projectRoot, resolvedPath)
		out.ProjectSummary = projectSummary
		out.ProjectErr = err
	}
	return out
}

func resolveChatProjectRoot(resolvedConfigPath string, serverFolder string) string {
	switch {
	case strings.TrimSpace(resolvedConfigPath) != "":
		return filepath.Clean(filepath.Dir(resolvedConfigPath))
	case strings.TrimSpace(serverFolder) != "":
		if filepath.IsAbs(serverFolder) {
			return filepath.Clean(serverFolder)
		}
		absPath, err := filepath.Abs(serverFolder)
		if err == nil {
			return filepath.Clean(absPath)
		}
		return filepath.Clean(serverFolder)
	default:
		return ""
	}
}

func loadContextFile(path string, limit int) (string, bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false, err
	}
	if len(b) <= limit {
		return string(b), false, nil
	}
	return string(b[:limit]), true, nil
}

func loadProjectSummary(projectRoot string, resolvedConfigPath string) (string, error) {
	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return "", err
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() == entries[j].IsDir() {
			return entries[i].Name() < entries[j].Name()
		}
		return entries[i].IsDir()
	})

	lines := []string{
		fmt.Sprintf("Project root: %s", projectRoot),
	}
	if resolvedConfigPath != "" {
		if relPath, err := filepath.Rel(projectRoot, resolvedConfigPath); err == nil {
			lines = append(lines, fmt.Sprintf("Primary config file: %s", filepath.ToSlash(relPath)))
		}
	}
	lines = append(lines, "Top-level project entries:")

	limit := len(entries)
	if limit > maxChatProjectEntries {
		limit = maxChatProjectEntries
	}
	for i := 0; i < limit; i++ {
		name := entries[i].Name()
		if entries[i].IsDir() {
			name += "/"
		}
		lines = append(lines, "- "+name)
	}
	if len(entries) > limit {
		lines = append(lines, fmt.Sprintf("- ... (%d more entries)", len(entries)-limit))
	}

	return strings.Join(lines, "\n"), nil
}

func buildChatInstructions(ctxInfo chatContext) (string, string) {
	return buildChatInstructionsWithBase(strings.TrimSpace(os.Getenv("OPENAI_CHAT_INSTRUCTIONS")), ctxInfo)
}

func buildChatInstructionsWithBase(baseInstructions string, ctxInfo chatContext) (string, string) {
	instructions := strings.TrimSpace(baseInstructions)
	if instructions == "" {
		instructions = defaultChatInstructions
	}

	var b strings.Builder
	b.WriteString(instructions)
	notices := make([]string, 0, 2)

	if ctxInfo.ProjectRoot != "" {
		notices = append(notices, fmt.Sprintf("Included project context for %s.", ctxInfo.ProjectLabel))
		b.WriteString("\n\nCurrent project workspace context follows.\n")
		if ctxInfo.ProjectSummary != "" {
			b.WriteString(ctxInfo.ProjectSummary)
		} else {
			b.WriteString("Project root: ")
			b.WriteString(ctxInfo.ProjectRoot)
		}
		b.WriteString("\n\nIf project tools are available, use them for file listing, file reads, edits, and git actions instead of guessing.")
		b.WriteString(" Paths passed to tools must stay inside the project root.")
		b.WriteString(" Never commit or push unless the user explicitly asks for that git action.")
	}

	if ctxInfo.Content != "" && ctxInfo.ResolvedPath != "" {
		notice := fmt.Sprintf("Included %s as request context.", ctxInfo.ContextLabel)
		if ctxInfo.Truncated {
			notice = fmt.Sprintf("Included %s as request context, truncated to %d bytes.", ctxInfo.ContextLabel, maxChatContextBytes)
		}
		notices = append(notices, notice)

		b.WriteString("\n\nCurrent project file context follows.\n")
		b.WriteString("Path: ")
		b.WriteString(ctxInfo.ResolvedPath)
		b.WriteString("\n\n")
		b.WriteString(ctxInfo.Content)
		if strings.EqualFold(filepath.Ext(ctxInfo.ResolvedPath), ".toml") {
			b.WriteString("\n\nIf the user wants config changes applied, edit this TOML directly with the available file tools.")
			b.WriteString(" After changing it, run process_config_and_review before your final answer so you can report render errors and image feedback from the updated config.")
		}
	}

	return b.String(), strings.Join(notices, " ")
}

func parseChatHistory(raw string) ([]models.ChatMessage, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	var history []models.ChatMessage
	if err := json.Unmarshal([]byte(raw), &history); err != nil {
		return nil, err
	}
	return normalizeChatHistory(history), nil
}

func normalizeChatHistory(history []models.ChatMessage) []models.ChatMessage {
	out := make([]models.ChatMessage, 0, len(history))
	for _, msg := range history {
		role := normalizeChatRole(msg.Role)
		content := strings.TrimSpace(msg.Content)
		if role == "" || content == "" {
			continue
		}
		out = append(out, models.ChatMessage{
			Role:    role,
			Content: content,
		})
	}
	if len(out) > maxChatHistoryMessages {
		out = out[len(out)-maxChatHistoryMessages:]
	}
	return out
}

func normalizeChatRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "user":
		return "user"
	case "assistant":
		return "assistant"
	default:
		return ""
	}
}

func normalizeProvider(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case providerDeepSeek:
		return providerDeepSeek
	case providerOpenAI:
		return providerOpenAI
	default:
		return ""
	}
}

func readRequestedProvider(r *http.Request) string {
	if provider := normalizeProvider(r.FormValue("provider")); provider != "" {
		return provider
	}
	return normalizeProvider(r.URL.Query().Get("provider"))
}

func resolveSelectedProvider(requestedProvider string, openAIAvailable bool, deepSeekAvailable bool) string {
	requestedProvider = normalizeProvider(requestedProvider)

	switch {
	case requestedProvider == providerOpenAI && openAIAvailable:
		return providerOpenAI
	case requestedProvider == providerDeepSeek && deepSeekAvailable:
		return providerDeepSeek
	case openAIAvailable && !deepSeekAvailable:
		return providerOpenAI
	case deepSeekAvailable && !openAIAvailable:
		return providerDeepSeek
	case openAIAvailable && deepSeekAvailable:
		if requestedProvider != "" {
			return requestedProvider
		}
		return providerOpenAI
	case requestedProvider != "":
		return requestedProvider
	default:
		return providerOpenAI
	}
}

func selectedProviderLabel(provider string, available bool) string {
	if !available {
		return "No API token configured"
	}
	switch provider {
	case providerDeepSeek:
		return "Using DEEPSEEK_API_TOKEN"
	default:
		return "Using OPENAI_API_KEY"
	}
}

func buildChatRouteURL(configEncoded string, serverFolderEncoded string, provider string) string {
	return buildChatURL("/chat", configEncoded, serverFolderEncoded, provider)
}

func buildChatRunURL(configEncoded string, serverFolderEncoded string, provider string) string {
	return buildChatURL("/api/chat/run", configEncoded, serverFolderEncoded, provider)
}

func buildChatURL(path string, configEncoded string, serverFolderEncoded string, provider string) string {
	values := url.Values{}
	if configEncoded != "" {
		values.Set("config", configEncoded)
	}
	if serverFolderEncoded != "" {
		values.Set("server_folder", serverFolderEncoded)
	}
	if provider = normalizeProvider(provider); provider != "" {
		values.Set("provider", provider)
	}
	encoded := values.Encode()
	if encoded == "" {
		return path
	}
	return path + "?" + encoded
}

func decodeMaybeBase64(raw string) string {
	if raw == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return raw
	}
	return string(decoded)
}

func readChatModel(r *http.Request, provider string) string {
	model := strings.TrimSpace(r.FormValue("model"))
	if model == "" {
		model = defaultModel(provider)
	}
	return model
}

func defaultModel(provider string) string {
	switch normalizeProvider(provider) {
	case providerDeepSeek:
		if model := strings.TrimSpace(os.Getenv("DEEPSEEK_MODEL")); model != "" {
			return model
		}
		return defaultDeepSeekModel
	default:
		if model := strings.TrimSpace(os.Getenv("OPENAI_MODEL")); model != "" {
			return model
		}
		return defaultOpenAIModel
	}
}

func baseURLForProvider(provider string) string {
	switch normalizeProvider(provider) {
	case providerDeepSeek:
		if baseURL := strings.TrimSpace(os.Getenv("DEEPSEEK_BASE_URL")); baseURL != "" {
			return strings.TrimRight(baseURL, "/")
		}
		return defaultDeepSeekBaseURL
	default:
		if baseURL := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL")); baseURL != "" {
			return strings.TrimRight(baseURL, "/")
		}
		return defaultOpenAIBaseURL
	}
}

func createProviderChatResponse(ctx context.Context, auth chatAuthConfig, model string, instructions string, history []models.ChatMessage) (string, error) {
	return createProviderChatResponseWithContext(ctx, auth, model, instructions, history, chatContext{})
}

func createProviderChatResponseWithContext(ctx context.Context, auth chatAuthConfig, model string, instructions string, history []models.ChatMessage, ctxInfo chatContext) (string, error) {
	result, err := createProviderChatResultWithContext(ctx, auth, model, instructions, history, ctxInfo)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(result.OutputText) == "" {
		switch normalizeProvider(auth.SelectedProvider) {
		case providerDeepSeek:
			return "", fmt.Errorf("deepseek response did not include text output")
		default:
			return "", fmt.Errorf("openai response did not include text output")
		}
	}
	return result.OutputText, nil
}

func createProviderChatResult(ctx context.Context, auth chatAuthConfig, model string, instructions string, history []models.ChatMessage) (chatProviderResult, error) {
	return createProviderChatResultWithContext(ctx, auth, model, instructions, history, chatContext{})
}

func createProviderChatResultWithContext(ctx context.Context, auth chatAuthConfig, model string, instructions string, history []models.ChatMessage, ctxInfo chatContext) (chatProviderResult, error) {
	return createProviderChatResultWithContextAndStatus(ctx, auth, model, instructions, history, ctxInfo, nil)
}

func createProviderChatResultWithContextAndStatus(ctx context.Context, auth chatAuthConfig, model string, instructions string, history []models.ChatMessage, ctxInfo chatContext, status chatStatusSink) (chatProviderResult, error) {
	switch normalizeProvider(auth.SelectedProvider) {
	case providerDeepSeek:
		reportChatStatus(status, "Waiting for DeepSeek response")
		result, err := createDeepSeekChatResult(ctx, auth.Token, baseURLForProvider(auth.SelectedProvider), model, instructions, history)
		if err != nil {
			return chatProviderResult{}, err
		}
		result.Provider = providerDeepSeek
		return result, nil
	default:
		reportChatStatus(status, "Waiting for OpenAI response")
		result, err := createOpenAIChatResultWithContextAndStatus(ctx, auth.Token, baseURLForProvider(auth.SelectedProvider), model, instructions, history, chatToolRuntime{
			CtxInfo: ctxInfo,
			Auth:    auth,
			Model:   model,
			Report:  status,
		})
		if err != nil {
			return chatProviderResult{}, err
		}
		result.Provider = providerOpenAI
		return result, nil
	}
}

func createOpenAIChatResponse(ctx context.Context, apiKey string, baseURL string, model string, instructions string, history []models.ChatMessage) (string, error) {
	result, err := createOpenAIChatResult(ctx, apiKey, baseURL, model, instructions, history)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(result.OutputText) == "" {
		return "", fmt.Errorf("openai response did not include text output")
	}
	return result.OutputText, nil
}

func createOpenAIChatResult(ctx context.Context, apiKey string, baseURL string, model string, instructions string, history []models.ChatMessage) (chatProviderResult, error) {
	return createOpenAIChatResultWithContext(ctx, apiKey, baseURL, model, instructions, history, chatContext{})
}

func createOpenAIChatResultWithContext(ctx context.Context, apiKey string, baseURL string, model string, instructions string, history []models.ChatMessage, ctxInfo chatContext) (chatProviderResult, error) {
	return createOpenAIChatResultWithContextAndStatus(ctx, apiKey, baseURL, model, instructions, history, chatToolRuntime{
		CtxInfo: ctxInfo,
		Auth: chatAuthConfig{
			SelectedProvider: providerOpenAI,
			Token:            apiKey,
		},
		Model: model,
	})
}

func createOpenAIChatResultWithContextAndStatus(ctx context.Context, apiKey string, baseURL string, model string, instructions string, history []models.ChatMessage, runtime chatToolRuntime) (chatProviderResult, error) {
	input, err := buildOpenAIInput(history)
	if err != nil {
		return chatProviderResult{}, err
	}
	if len(input) == 0 {
		return chatProviderResult{}, fmt.Errorf("no chat messages to send")
	}

	tools := buildOpenAIProjectTools(runtime.CtxInfo)
	allToolCalls := make([]chatToolCall, 0)

	for round := 0; round < maxChatToolRounds; round++ {
		if round > 0 {
			reportChatStatus(runtime.Report, fmt.Sprintf("Continuing tool loop (%d/%d)", round+1, maxChatToolRounds))
		}
		reqBody := openAIResponsesRequest{
			Model:        model,
			Instructions: instructions,
			Input:        input,
		}
		if len(tools) > 0 {
			reqBody.Tools = tools
			reqBody.ToolChoice = "auto"
			reqBody.ParallelToolCalls = false
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return chatProviderResult{}, fmt.Errorf("encode openai request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/v1/responses", bytes.NewReader(jsonBody))
		if err != nil {
			return chatProviderResult{}, fmt.Errorf("build openai request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 90 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return chatProviderResult{}, fmt.Errorf("send openai request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return chatProviderResult{}, fmt.Errorf("read openai response: %w", err)
		}

		if resp.StatusCode >= http.StatusBadRequest {
			return chatProviderResult{}, fmt.Errorf("openai request failed: %s", formatOpenAIError(resp.StatusCode, body))
		}

		result, err := parseOpenAIChatResult(body)
		if err != nil {
			return chatProviderResult{}, err
		}
		if result.Model == "" {
			result.Model = model
		}
		allToolCalls = append(allToolCalls, result.ToolCalls...)

		if len(result.ToolCalls) == 0 {
			result.ToolCalls = allToolCalls
			return result, nil
		}
		if len(tools) == 0 {
			result.ToolCalls = allToolCalls
			return result, nil
		}

		reportChatStatus(runtime.Report, fmt.Sprintf("Running %d tool action(s)", len(result.ToolCalls)))
		toolOutputs, err := executeChatToolCalls(ctx, runtime, result.ToolCalls)
		if err != nil {
			return chatProviderResult{}, err
		}
		input = append(input, result.OutputItems...)
		input = append(input, toolOutputs...)
	}

	return chatProviderResult{}, fmt.Errorf("openai tool execution exceeded %d rounds", maxChatToolRounds)
}

func buildOpenAIInput(history []models.ChatMessage) ([]json.RawMessage, error) {
	input := make([]json.RawMessage, 0, len(history))
	for _, msg := range normalizeChatHistory(history) {
		item, err := json.Marshal(openAIInputMessage{
			Role: msg.Role,
			Content: []openAITextInput{
				{
					Type: "input_text",
					Text: msg.Content,
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("encode openai input message: %w", err)
		}
		input = append(input, json.RawMessage(item))
	}
	return input, nil
}

func chatProjectToolsEnabled(provider string, ctxInfo chatContext) bool {
	return normalizeProvider(provider) == providerOpenAI && strings.TrimSpace(ctxInfo.ProjectRoot) != ""
}

func chatConfigAutomationEnabled(auth chatAuthConfig, ctxInfo chatContext) bool {
	if !chatProjectToolsEnabled(auth.SelectedProvider, ctxInfo) {
		return false
	}
	if strings.TrimSpace(ctxInfo.ResolvedPath) == "" {
		return false
	}
	return strings.EqualFold(filepath.Ext(ctxInfo.ResolvedPath), ".toml")
}

func buildOpenAIProjectTools(ctxInfo chatContext) []openAITool {
	if strings.TrimSpace(ctxInfo.ProjectRoot) == "" {
		return nil
	}

	stringSchema := map[string]any{"type": "string"}
	return []openAITool{
		{
			Type:        "function",
			Name:        "list_project_files",
			Description: "List files and folders inside the current project root. Use relative paths and small depths before reading or editing files.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":  stringSchema,
					"depth": map[string]any{"type": "integer", "minimum": 0, "maximum": 4},
				},
				"additionalProperties": false,
			},
		},
		{
			Type:        "function",
			Name:        "read_project_file",
			Description: "Read a UTF-8 text file from the current project root. Use this before making edits.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":      stringSchema,
					"offset":    map[string]any{"type": "integer", "minimum": 0},
					"max_bytes": map[string]any{"type": "integer", "minimum": 1, "maximum": maxChatReadBytes},
				},
				"required":             []string{"path"},
				"additionalProperties": false,
			},
		},
		{
			Type:        "function",
			Name:        "edit_project_file",
			Description: "Replace text inside a project file using exact old_text and new_text values. Use this for targeted edits within the project root.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":        stringSchema,
					"old_text":    stringSchema,
					"new_text":    stringSchema,
					"replace_all": map[string]any{"type": "boolean"},
				},
				"required":             []string{"path", "old_text", "new_text"},
				"additionalProperties": false,
			},
		},
		{
			Type:        "function",
			Name:        "write_project_file",
			Description: "Create or fully overwrite a text file inside the project root.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":    stringSchema,
					"content": stringSchema,
				},
				"required":             []string{"path", "content"},
				"additionalProperties": false,
			},
		},
		buildProcessConfigTool(),
		{
			Type:        "function",
			Name:        "git_status",
			Description: "Show git status for the current project root.",
			Strict:      true,
			Parameters: map[string]any{
				"type":                 "object",
				"properties":           map[string]any{},
				"additionalProperties": false,
			},
		},
		{
			Type:        "function",
			Name:        "git_diff",
			Description: "Show git diff for the current project root, optionally scoped to one relative path.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":   stringSchema,
					"cached": map[string]any{"type": "boolean"},
				},
				"additionalProperties": false,
			},
		},
		{
			Type:        "function",
			Name:        "git_commit",
			Description: "Stage and commit project-root changes with the supplied commit message. Only use this when the user explicitly asks for a commit.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"message": map[string]any{"type": "string", "minLength": 1},
					"paths": map[string]any{
						"type":  "array",
						"items": stringSchema,
					},
				},
				"required":             []string{"message"},
				"additionalProperties": false,
			},
		},
		{
			Type:        "function",
			Name:        "git_push",
			Description: "Push the current branch or the supplied remote and branch. Only use this when the user explicitly asks for a push.",
			Strict:      true,
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"remote":       stringSchema,
					"branch":       stringSchema,
					"set_upstream": map[string]any{"type": "boolean"},
				},
				"additionalProperties": false,
			},
		},
	}
}

func executeChatToolCalls(ctx context.Context, runtime chatToolRuntime, calls []chatToolCall) ([]json.RawMessage, error) {
	outputs := make([]json.RawMessage, 0, len(calls))
	for _, call := range calls {
		output, err := executeChatToolCall(ctx, runtime, call)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
}

func executeChatToolCall(ctx context.Context, runtime chatToolRuntime, call chatToolCall) (json.RawMessage, error) {
	callID := strings.TrimSpace(call.CallID)
	if callID == "" {
		callID = strings.TrimSpace(call.ID)
	}
	if callID == "" {
		return nil, fmt.Errorf("tool call %q did not include a call id", call.Name)
	}

	reportChatStatus(runtime.Report, describeChatToolCall(call))
	payload, err := runProjectTool(ctx, runtime, call)
	if err != nil {
		payload = map[string]any{
			"ok":    false,
			"tool":  call.Name,
			"error": err.Error(),
		}
		reportChatStatus(runtime.Report, fmt.Sprintf("Tool %s failed", call.Name))
	} else {
		reportChatStatus(runtime.Report, fmt.Sprintf("Tool %s completed", call.Name))
	}
	return marshalFunctionCallOutput(callID, payload)
}

func marshalFunctionCallOutput(callID string, payload any) (json.RawMessage, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode tool payload: %w", err)
	}
	item, err := json.Marshal(openAIFunctionCallOutput{
		Type:   "function_call_output",
		CallID: callID,
		Output: string(body),
	})
	if err != nil {
		return nil, fmt.Errorf("encode tool output: %w", err)
	}
	return json.RawMessage(item), nil
}

func runProjectTool(ctx context.Context, runtime chatToolRuntime, call chatToolCall) (any, error) {
	projectRoot := strings.TrimSpace(runtime.CtxInfo.ProjectRoot)
	if projectRoot == "" {
		return nil, fmt.Errorf("no project root is attached to this chat")
	}

	switch call.Name {
	case "list_project_files":
		var args struct {
			Path  string `json:"path"`
			Depth int    `json:"depth"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return listProjectFilesTool(projectRoot, args.Path, args.Depth)
	case "read_project_file":
		var args struct {
			Path     string `json:"path"`
			Offset   int    `json:"offset"`
			MaxBytes int    `json:"max_bytes"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return readProjectFileTool(projectRoot, args.Path, args.Offset, args.MaxBytes)
	case "edit_project_file":
		var args struct {
			Path       string `json:"path"`
			OldText    string `json:"old_text"`
			NewText    string `json:"new_text"`
			ReplaceAll bool   `json:"replace_all"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return editProjectFileTool(projectRoot, args.Path, args.OldText, args.NewText, args.ReplaceAll)
	case "write_project_file":
		var args struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return writeProjectFileTool(projectRoot, args.Path, args.Content)
	case "process_config_and_review":
		var args struct {
			Path         string `json:"path"`
			ReviewImages *bool  `json:"review_images"`
			MaxImages    int    `json:"max_images"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		reviewImages := true
		if args.ReviewImages != nil {
			reviewImages = *args.ReviewImages
		}
		return processConfigAndReviewTool(ctx, runtime, args.Path, reviewImages, args.MaxImages)
	case "git_status":
		return gitStatusTool(ctx, projectRoot)
	case "git_diff":
		var args struct {
			Path   string `json:"path"`
			Cached bool   `json:"cached"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return gitDiffTool(ctx, projectRoot, args.Path, args.Cached)
	case "git_commit":
		var args struct {
			Message string   `json:"message"`
			Paths   []string `json:"paths"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return gitCommitTool(ctx, projectRoot, args.Message, args.Paths)
	case "git_push":
		var args struct {
			Remote      string `json:"remote"`
			Branch      string `json:"branch"`
			SetUpstream bool   `json:"set_upstream"`
		}
		if err := decodeToolArguments(call.Arguments, &args); err != nil {
			return nil, err
		}
		return gitPushTool(ctx, projectRoot, args.Remote, args.Branch, args.SetUpstream)
	default:
		return nil, fmt.Errorf("unsupported tool %q", call.Name)
	}
}

func describeChatToolCall(call chatToolCall) string {
	type pathArg struct {
		Path string `json:"path"`
	}

	switch call.Name {
	case "list_project_files":
		var args pathArg
		_ = decodeToolArguments(call.Arguments, &args)
		if strings.TrimSpace(args.Path) == "" {
			return "Listing project files"
		}
		return "Listing project files in " + args.Path
	case "read_project_file":
		var args pathArg
		_ = decodeToolArguments(call.Arguments, &args)
		return "Reading " + args.Path
	case "edit_project_file":
		var args pathArg
		_ = decodeToolArguments(call.Arguments, &args)
		return "Applying edits to " + args.Path
	case "write_project_file":
		var args pathArg
		_ = decodeToolArguments(call.Arguments, &args)
		return "Writing " + args.Path
	case "process_config_and_review":
		var args pathArg
		_ = decodeToolArguments(call.Arguments, &args)
		return "Processing " + args.Path
	case "git_status":
		return "Checking git status"
	case "git_diff":
		return "Inspecting git diff"
	case "git_commit":
		return "Creating git commit"
	case "git_push":
		return "Pushing git branch"
	default:
		return "Running " + call.Name
	}
}

func decodeToolArguments(raw string, dest any) error {
	if strings.TrimSpace(raw) == "" {
		raw = "{}"
	}
	if err := json.Unmarshal([]byte(raw), dest); err != nil {
		return fmt.Errorf("decode tool arguments: %w", err)
	}
	return nil
}

func listProjectFilesTool(projectRoot string, requestedPath string, depth int) (map[string]any, error) {
	if depth <= 0 {
		depth = 2
	}
	if depth > 4 {
		depth = 4
	}

	targetAbs, targetRel, err := resolveProjectPath(projectRoot, requestedPath, true)
	if err != nil {
		return nil, err
	}

	entries := make([]map[string]any, 0)
	truncated := false
	var walk func(string, string, int) error
	walk = func(currentAbs string, currentRel string, remaining int) error {
		dirEntries, err := os.ReadDir(currentAbs)
		if err != nil {
			return err
		}
		sort.Slice(dirEntries, func(i, j int) bool {
			if dirEntries[i].IsDir() == dirEntries[j].IsDir() {
				return dirEntries[i].Name() < dirEntries[j].Name()
			}
			return dirEntries[i].IsDir()
		})
		for _, entry := range dirEntries {
			if len(entries) >= maxChatProjectEntries {
				truncated = true
				return nil
			}

			childRel := entry.Name()
			if currentRel != "." {
				childRel = filepath.Join(currentRel, childRel)
			}
			entryType := "file"
			if entry.IsDir() {
				entryType = "dir"
			}
			entries = append(entries, map[string]any{
				"path": filepath.ToSlash(childRel),
				"type": entryType,
			})
			if entry.IsDir() && remaining > 0 {
				if err := walk(filepath.Join(currentAbs, entry.Name()), childRel, remaining-1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := walk(targetAbs, targetRel, depth-1); err != nil {
		return nil, err
	}

	return map[string]any{
		"ok":           true,
		"project_root": projectRoot,
		"path":         filepath.ToSlash(targetRel),
		"depth":        depth,
		"entries":      entries,
		"truncated":    truncated,
	}, nil
}

func readProjectFileTool(projectRoot string, requestedPath string, offset int, maxBytes int) (map[string]any, error) {
	targetAbs, targetRel, err := resolveProjectPath(projectRoot, requestedPath, false)
	if err != nil {
		return nil, err
	}
	if maxBytes <= 0 || maxBytes > maxChatReadBytes {
		maxBytes = maxChatReadBytes
	}
	if offset < 0 {
		offset = 0
	}

	body, err := os.ReadFile(targetAbs)
	if err != nil {
		return nil, err
	}
	if offset > len(body) {
		offset = len(body)
	}

	end := offset + maxBytes
	truncated := false
	if end > len(body) {
		end = len(body)
	} else if end < len(body) {
		truncated = true
	}

	return map[string]any{
		"ok":         true,
		"path":       filepath.ToSlash(targetRel),
		"offset":     offset,
		"content":    string(body[offset:end]),
		"truncated":  truncated,
		"total_size": len(body),
	}, nil
}

func editProjectFileTool(projectRoot string, requestedPath string, oldText string, newText string, replaceAll bool) (map[string]any, error) {
	targetAbs, targetRel, err := resolveProjectPath(projectRoot, requestedPath, false)
	if err != nil {
		return nil, err
	}
	if oldText == "" {
		return nil, fmt.Errorf("old_text must be non-empty")
	}

	body, err := os.ReadFile(targetAbs)
	if err != nil {
		return nil, err
	}

	content := string(body)
	if !strings.Contains(content, oldText) {
		return nil, fmt.Errorf("old_text was not found in %s", filepath.ToSlash(targetRel))
	}

	replacements := 1
	if replaceAll {
		replacements = -1
	}
	updated := strings.Replace(content, oldText, newText, replacements)
	if updated == content {
		return nil, fmt.Errorf("edit produced no change for %s", filepath.ToSlash(targetRel))
	}

	info, err := os.Stat(targetAbs)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(targetAbs, []byte(updated), info.Mode().Perm()); err != nil {
		return nil, err
	}
	if shouldJournalProjectWrite(targetRel) {
		if err := pkg.RecordUpdateJournal(filepath.Dir(targetAbs), filepath.Base(targetAbs), "file edited", "file saved", fmt.Sprintf("path=%s replace_all=%t", filepath.ToSlash(targetRel), replaceAll), true); err != nil {
			pkg.LogWarnf("update journal write failed: %v", err)
		}
	}

	count := strings.Count(content, oldText)
	if !replaceAll && count > 1 {
		count = 1
	}
	return map[string]any{
		"ok":            true,
		"path":          filepath.ToSlash(targetRel),
		"replacements":  count,
		"bytes_written": len(updated),
	}, nil
}

func writeProjectFileTool(projectRoot string, requestedPath string, content string) (map[string]any, error) {
	targetAbs, targetRel, err := resolveProjectPath(projectRoot, requestedPath, false)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		return nil, err
	}

	mode := fs.FileMode(0o644)
	if info, err := os.Stat(targetAbs); err == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(targetAbs, []byte(content), mode); err != nil {
		return nil, err
	}
	if shouldJournalProjectWrite(targetRel) {
		if err := pkg.RecordUpdateJournal(filepath.Dir(targetAbs), filepath.Base(targetAbs), "file written", "file saved", fmt.Sprintf("path=%s", filepath.ToSlash(targetRel)), true); err != nil {
			pkg.LogWarnf("update journal write failed: %v", err)
		}
	}

	return map[string]any{
		"ok":            true,
		"path":          filepath.ToSlash(targetRel),
		"bytes_written": len(content),
	}, nil
}

func shouldJournalProjectWrite(targetRel string) bool {
	base := filepath.Base(targetRel)
	return strings.EqualFold(filepath.Ext(targetRel), ".toml") || strings.EqualFold(base, "config.toml")
}

func gitStatusTool(ctx context.Context, projectRoot string) (map[string]any, error) {
	output, err := runProjectCommand(ctx, projectRoot, 20*time.Second, "git", "status", "--short", "--branch", "--", ".")
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":     true,
		"output": output,
	}, nil
}

func gitDiffTool(ctx context.Context, projectRoot string, requestedPath string, cached bool) (map[string]any, error) {
	args := []string{"diff"}
	if cached {
		args = append(args, "--cached")
	}
	args = append(args, "--")
	if strings.TrimSpace(requestedPath) == "" {
		args = append(args, ".")
	} else {
		_, targetRel, err := resolveProjectPath(projectRoot, requestedPath, false)
		if err != nil {
			return nil, err
		}
		args = append(args, targetRel)
	}

	output, err := runProjectCommand(ctx, projectRoot, 20*time.Second, "git", args...)
	if err != nil {
		return nil, err
	}
	output, truncated := trimToolOutput(output, maxChatDiffBytes)
	return map[string]any{
		"ok":        true,
		"output":    output,
		"cached":    cached,
		"truncated": truncated,
	}, nil
}

func gitCommitTool(ctx context.Context, projectRoot string, message string, requestedPaths []string) (map[string]any, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("commit message must be non-empty")
	}

	if err := ensureNoStagedChangesOutsideProject(ctx, projectRoot); err != nil {
		return nil, err
	}

	addArgs := []string{"add", "-A", "--"}
	if len(requestedPaths) == 0 {
		addArgs = append(addArgs, ".")
	} else {
		for _, requestedPath := range requestedPaths {
			_, targetRel, err := resolveProjectPath(projectRoot, requestedPath, false)
			if err != nil {
				return nil, err
			}
			addArgs = append(addArgs, targetRel)
		}
	}
	if _, err := runProjectCommand(ctx, projectRoot, 20*time.Second, "git", addArgs...); err != nil {
		return nil, err
	}
	if err := ensureNoStagedChangesOutsideProject(ctx, projectRoot); err != nil {
		return nil, err
	}

	stagedOutput, err := runProjectCommand(ctx, projectRoot, 10*time.Second, "git", "diff", "--cached", "--name-only", "--", ".")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(stagedOutput) == "" {
		return nil, fmt.Errorf("no staged project changes to commit")
	}

	commitOutput, err := runProjectCommand(ctx, projectRoot, 30*time.Second, "git", "commit", "-m", message)
	if err != nil {
		return nil, err
	}
	head, err := runProjectCommand(ctx, projectRoot, 10*time.Second, "git", "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"ok":      true,
		"commit":  strings.TrimSpace(head),
		"output":  commitOutput,
		"message": message,
	}, nil
}

func gitPushTool(ctx context.Context, projectRoot string, remote string, branch string, setUpstream bool) (map[string]any, error) {
	remote = strings.TrimSpace(remote)
	branch = strings.TrimSpace(branch)
	if setUpstream && (remote == "" || branch == "") {
		return nil, fmt.Errorf("set_upstream requires both remote and branch")
	}
	if remote == "" && branch != "" {
		return nil, fmt.Errorf("branch cannot be set without remote")
	}

	args := []string{"push"}
	if setUpstream {
		args = append(args, "-u")
	}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}

	output, err := runProjectCommand(ctx, projectRoot, 2*time.Minute, "git", args...)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":       true,
		"output":   output,
		"remote":   remote,
		"branch":   branch,
		"upstream": setUpstream,
	}, nil
}

func resolveProjectPath(projectRoot string, requestedPath string, allowRoot bool) (string, string, error) {
	projectRoot = filepath.Clean(strings.TrimSpace(projectRoot))
	if projectRoot == "" {
		return "", "", fmt.Errorf("project root is empty")
	}

	cleanRequested := strings.TrimSpace(requestedPath)
	if cleanRequested == "" {
		if allowRoot {
			return projectRoot, ".", nil
		}
		return "", "", fmt.Errorf("path is required")
	}

	var candidate string
	if filepath.IsAbs(cleanRequested) {
		candidate = filepath.Clean(cleanRequested)
	} else {
		candidate = filepath.Clean(filepath.Join(projectRoot, cleanRequested))
	}

	rel, err := filepath.Rel(projectRoot, candidate)
	if err != nil {
		return "", "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("path %q is outside the project root", requestedPath)
	}
	if rel == "." && !allowRoot {
		return "", "", fmt.Errorf("path must point to a file or subdirectory inside the project root")
	}
	return candidate, rel, nil
}

func runProjectCommand(ctx context.Context, projectRoot string, timeout time.Duration, name string, args ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, name, args...)
	cmd.Dir = projectRoot
	outputBytes, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(outputBytes))
	output, _ = trimToolOutput(output, maxChatOutputBytes)

	if cmdCtx.Err() == context.DeadlineExceeded {
		return output, fmt.Errorf("%s timed out", name)
	}
	if err != nil {
		if output == "" {
			output = err.Error()
		}
		return output, fmt.Errorf("%s %s failed: %s", name, strings.Join(args, " "), output)
	}
	return output, nil
}

func trimToolOutput(output string, limit int) (string, bool) {
	if limit <= 0 || len(output) <= limit {
		return output, false
	}
	return output[:limit], true
}

func ensureNoStagedChangesOutsideProject(ctx context.Context, projectRoot string) error {
	repoRoot, err := runProjectCommand(ctx, projectRoot, 10*time.Second, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	repoRoot = strings.TrimSpace(repoRoot)

	projectRel, err := filepath.Rel(repoRoot, projectRoot)
	if err != nil {
		return err
	}
	projectRel = filepath.ToSlash(filepath.Clean(projectRel))

	staged, err := runProjectCommand(ctx, projectRoot, 10*time.Second, "git", "diff", "--cached", "--name-only")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(staged, "\n") {
		line = filepath.ToSlash(strings.TrimSpace(line))
		if line == "" {
			continue
		}
		if projectRel == "." {
			continue
		}
		if line != projectRel && !strings.HasPrefix(line, projectRel+"/") {
			return fmt.Errorf("staged changes outside the current project root would be included in the commit: %s", line)
		}
	}
	return nil
}

func parseOpenAIChatResult(body []byte) (chatProviderResult, error) {
	var parsed openAIResponsesResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return chatProviderResult{}, fmt.Errorf("decode openai response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return chatProviderResult{}, fmt.Errorf("openai error: %s", parsed.Error.Message)
	}

	var rawEnvelope openAIResponsesRawEnvelope
	if err := json.Unmarshal(body, &rawEnvelope); err != nil {
		return chatProviderResult{}, fmt.Errorf("decode openai raw response: %w", err)
	}

	result := chatProviderResult{
		Model:       parsed.Model,
		OutputText:  extractOutputText(parsed),
		ToolCalls:   extractOutputToolCalls(rawEnvelope.Output),
		RawResponse: append(json.RawMessage(nil), body...),
		OutputItems: append([]json.RawMessage(nil), rawEnvelope.Output...),
	}
	if strings.TrimSpace(result.OutputText) == "" && len(result.ToolCalls) == 0 {
		return chatProviderResult{}, fmt.Errorf("openai response did not include text output or tool calls")
	}
	return result, nil
}

func createDeepSeekChatResponse(ctx context.Context, apiKey string, baseURL string, model string, instructions string, history []models.ChatMessage) (string, error) {
	result, err := createDeepSeekChatResult(ctx, apiKey, baseURL, model, instructions, history)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(result.OutputText) == "" {
		return "", fmt.Errorf("deepseek response did not include text output")
	}
	return result.OutputText, nil
}

func createDeepSeekChatResult(ctx context.Context, apiKey string, baseURL string, model string, instructions string, history []models.ChatMessage) (chatProviderResult, error) {
	messages := make([]deepSeekChatMessage, 0, len(history)+1)
	if strings.TrimSpace(instructions) != "" {
		messages = append(messages, deepSeekChatMessage{
			Role:    "system",
			Content: instructions,
		})
	}
	for _, msg := range normalizeChatHistory(history) {
		messages = append(messages, deepSeekChatMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	if len(messages) == 0 {
		return chatProviderResult{}, fmt.Errorf("no chat messages to send")
	}

	reqBody := deepSeekChatCompletionsRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return chatProviderResult{}, fmt.Errorf("encode deepseek request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(baseURL, "/")+"/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return chatProviderResult{}, fmt.Errorf("build deepseek request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return chatProviderResult{}, fmt.Errorf("send deepseek request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return chatProviderResult{}, fmt.Errorf("read deepseek response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return chatProviderResult{}, fmt.Errorf("deepseek request failed: %s", formatOpenAIError(resp.StatusCode, body))
	}

	var parsed deepSeekChatCompletionsResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return chatProviderResult{}, fmt.Errorf("decode deepseek response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return chatProviderResult{}, fmt.Errorf("deepseek error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return chatProviderResult{}, fmt.Errorf("deepseek response did not include choices")
	}

	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return chatProviderResult{}, fmt.Errorf("deepseek response did not include text output")
	}
	return chatProviderResult{
		Model:       parsed.Model,
		OutputText:  text,
		RawResponse: append(json.RawMessage(nil), body...),
	}, nil
}

func formatOpenAIError(statusCode int, body []byte) string {
	var envelope openAIErrorEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Error != nil && envelope.Error.Message != "" {
		return fmt.Sprintf("%d %s", statusCode, envelope.Error.Message)
	}

	msg := strings.TrimSpace(string(body))
	if msg == "" {
		return fmt.Sprintf("%d", statusCode)
	}
	return fmt.Sprintf("%d %s", statusCode, msg)
}

func extractOutputText(resp openAIResponsesResponse) string {
	parts := make([]string, 0, len(resp.Output))
	for _, item := range resp.Output {
		if item.Type != "message" {
			continue
		}
		for _, content := range item.Content {
			if content.Type != "output_text" {
				continue
			}
			text := strings.TrimSpace(content.Text)
			if text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.Join(parts, "\n\n")
}

func extractOutputToolCalls(items []json.RawMessage) []chatToolCall {
	out := make([]chatToolCall, 0)
	for _, item := range items {
		var parsed chatToolCall
		if err := json.Unmarshal(item, &parsed); err != nil {
			continue
		}
		if !strings.HasSuffix(parsed.Type, "_call") {
			continue
		}
		parsed.Raw = append(json.RawMessage(nil), item...)
		out = append(out, parsed)
	}
	return out
}
