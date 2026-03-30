package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/acp"
	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/approvals"
	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/ascbin"
	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/environment"
	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/settings"
	"github.com/rudrankriyam/App-Store-Connect-CLI/apps/studio/internal/studio/threads"
)

type App struct {
	ctx         context.Context
	rootDir     string
	settings    *settings.Store
	threads     *threads.Store
	approvals   *approvals.Queue
	environment *environment.Service

	mu       sync.Mutex
	sessions map[string]*threadSession
}

type threadSession struct {
	client    *acp.Client
	sessionID string
}

type BootstrapData struct {
	AppName      string                    `json:"appName"`
	Tagline      string                    `json:"tagline"`
	GeneratedAt  time.Time                 `json:"generatedAt"`
	Sections     []WorkspaceSection        `json:"sections"`
	Settings     settings.StudioSettings   `json:"settings"`
	Presets      []settings.ProviderPreset `json:"presets"`
	Environment  environment.Snapshot      `json:"environment"`
	Threads      []threads.Thread          `json:"threads"`
	Approvals    []approvals.Action        `json:"approvals"`
	WindowFlavor string                    `json:"windowFlavor"`
}

type WorkspaceSection struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type PromptRequest struct {
	ThreadID string `json:"threadId"`
	Prompt   string `json:"prompt"`
}

type PromptResponse struct {
	Thread threads.Thread `json:"thread"`
}

type ApprovalRequest struct {
	ThreadID        string   `json:"threadId"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	CommandPreview  []string `json:"commandPreview"`
	MutationSurface string   `json:"mutationSurface"`
}

type ResolutionResponse struct {
	ascbin.Resolution
	AvailablePresets []settings.ProviderPreset `json:"availablePresets"`
}

type AuthStatus struct {
	Authenticated bool   `json:"authenticated"`
	Storage       string `json:"storage"`
	Profile       string `json:"profile"`
	RawOutput     string `json:"rawOutput"`
}

type AppInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Subtitle string `json:"subtitle"`
	BundleID string `json:"bundleId"`
	Platform string `json:"platform"`
	SKU      string `json:"sku"`
}

type ListAppsResponse struct {
	Apps  []AppInfo `json:"apps"`
	Error string    `json:"error,omitempty"`
}

type AppVersion struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	Version  string `json:"version"`
	State    string `json:"state"`
}

type AppDetail struct {
	ID            string       `json:"id"`
	Name          string       `json:"name"`
	Subtitle      string       `json:"subtitle"`
	BundleID      string       `json:"bundleId"`
	SKU           string       `json:"sku"`
	PrimaryLocale string       `json:"primaryLocale"`
	Versions      []AppVersion `json:"versions"`
	Error         string       `json:"error,omitempty"`
}

type ASCCommandResponse struct {
	Data  string `json:"data"`
	Error string `json:"error,omitempty"`
}

type AppLocalization struct {
	LocalizationID  string `json:"localizationId"`
	Locale          string `json:"locale"`
	Description     string `json:"description"`
	Keywords        string `json:"keywords"`
	WhatsNew        string `json:"whatsNew"`
	PromotionalText string `json:"promotionalText"`
	SupportURL      string `json:"supportUrl"`
	MarketingURL    string `json:"marketingUrl"`
}

type VersionMetadataResponse struct {
	Localizations []AppLocalization `json:"localizations"`
	Error         string            `json:"error,omitempty"`
}

type AppScreenshot struct {
	ThumbnailURL string `json:"thumbnailUrl"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type ScreenshotSet struct {
	DisplayType string          `json:"displayType"`
	Screenshots []AppScreenshot `json:"screenshots"`
}

type ScreenshotsResponse struct {
	Sets  []ScreenshotSet `json:"sets"`
	Error string          `json:"error,omitempty"`
}

func NewApp() (*App, error) {
	rootDir, err := settings.DefaultRoot()
	if err != nil {
		return nil, err
	}

	return &App{
		rootDir:     rootDir,
		settings:    settings.NewStore(rootDir),
		threads:     threads.NewStore(rootDir),
		approvals:   approvals.NewQueue(),
		environment: environment.NewService(),
		sessions:    make(map[string]*threadSession),
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for key, session := range a.sessions {
		if session != nil && session.client != nil {
			_ = session.client.Close()
		}
		delete(a.sessions, key)
	}
}

func (a *App) Bootstrap() (BootstrapData, error) {
	cfg, err := a.settings.Load()
	if err != nil {
		return BootstrapData{}, err
	}

	snapshot, err := a.environment.Snapshot()
	if err != nil {
		return BootstrapData{}, err
	}

	existingThreads, err := a.threads.LoadAll()
	if err != nil {
		return BootstrapData{}, err
	}

	return BootstrapData{
		AppName:      "ASC Studio",
		Tagline:      "The glassy desktop workspace for App Store Connect, powered by asc.",
		GeneratedAt:  time.Now().UTC(),
		Sections:     defaultSections(),
		Settings:     cfg,
		Presets:      settings.DefaultPresets(),
		Environment:  snapshot,
		Threads:      existingThreads,
		Approvals:    a.approvals.Pending(),
		WindowFlavor: "translucent",
	}, nil
}

func (a *App) GetSettings() (settings.StudioSettings, error) {
	return a.settings.Load()
}

func (a *App) SaveSettings(next settings.StudioSettings) (settings.StudioSettings, error) {
	next.Normalize()
	if err := a.settings.Save(next); err != nil {
		return settings.StudioSettings{}, err
	}
	return a.settings.Load()
}

func (a *App) CheckAuthStatus() (AuthStatus, error) {
	ascPath, err := a.resolveASCPath()
	if err != nil {
		return AuthStatus{RawOutput: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ascPath, "auth", "status", "--output", "json")
	cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
	out, err := cmd.CombinedOutput()
	output := strings.TrimSpace(string(out))

	status := AuthStatus{RawOutput: output}

	if err != nil {
		status.Authenticated = false
		return status, nil
	}

	// Exit 0 means credentials exist. Try to parse JSON output.
	status.Authenticated = true

	var jsonStatus struct {
		StorageBackend  string `json:"storageBackend"`
		StorageLocation string `json:"storageLocation"`
		Credentials     []struct {
			Name      string `json:"name"`
			KeyID     string `json:"keyId"`
			IsDefault bool   `json:"isDefault"`
		} `json:"credentials"`
	}
	if json.Unmarshal([]byte(output), &jsonStatus) == nil {
		status.Storage = jsonStatus.StorageBackend
		for _, cred := range jsonStatus.Credentials {
			if cred.IsDefault {
				status.Profile = cred.Name
				break
			}
		}
		if status.Profile == "" && len(jsonStatus.Credentials) > 0 {
			status.Profile = jsonStatus.Credentials[0].Name
		}
	}

	return status, nil
}

func (a *App) ListApps() (ListAppsResponse, error) {
	ascPath, err := a.resolveASCPath()
	if err != nil {
		return ListAppsResponse{Error: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ascPath, "apps", "list", "--output", "json")
	cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ListAppsResponse{Error: strings.TrimSpace(string(out))}, nil
	}

	// asc apps list --output json returns {"data":[...]} or a bare array
	type rawApp struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			Name     string `json:"name"`
			BundleID string `json:"bundleId"`
			SKU      string `json:"sku"`
		} `json:"attributes"`
	}

	var rawApps []rawApp

	// Try {"data":[...]} envelope first
	var envelope struct {
		Data []rawApp `json:"data"`
	}
	if err := json.Unmarshal(out, &envelope); err == nil && len(envelope.Data) > 0 {
		rawApps = envelope.Data
	} else if err := json.Unmarshal(out, &rawApps); err != nil {
		return ListAppsResponse{Error: "Failed to parse apps list: " + err.Error()}, nil
	}

	apps := make([]AppInfo, len(rawApps))
	for i, raw := range rawApps {
		apps[i] = AppInfo{
			ID:       raw.ID,
			Name:     raw.Attributes.Name,
			BundleID: raw.Attributes.BundleID,
			SKU:      raw.Attributes.SKU,
		}
	}

	// Fetch subtitles concurrently (best-effort; failures are silently skipped)
	subtitleCtx, subtitleCancel := context.WithTimeout(a.contextOrBackground(), 20*time.Second)
	defer subtitleCancel()

	type subtitleResult struct {
		index    int
		subtitle string
	}
	results := make(chan subtitleResult, len(apps))
	for i, app := range apps {
		go func(idx int, appID string) {
			subtitle := a.fetchSubtitle(subtitleCtx, ascPath, appID)
			results <- subtitleResult{index: idx, subtitle: subtitle}
		}(i, app.ID)
	}
	for range apps {
		r := <-results
		apps[r.index].Subtitle = r.subtitle
	}

	return ListAppsResponse{Apps: apps}, nil
}

func (a *App) fetchSubtitle(ctx context.Context, ascPath, appID string) string {
	cmd := exec.CommandContext(ctx, ascPath, "localizations", "list",
		"--app", appID, "--type", "app-info", "--locale", "en-US", "--output", "json")
	cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}

	type locAttrs struct {
		Subtitle string `json:"subtitle"`
	}
	type locItem struct {
		Attributes locAttrs `json:"attributes"`
	}
	var envelope struct {
		Data []locItem `json:"data"`
	}
	if json.Unmarshal(out, &envelope) != nil || len(envelope.Data) == 0 {
		return ""
	}
	return envelope.Data[0].Attributes.Subtitle
}

// RunASCCommand runs an arbitrary asc CLI command and returns the raw JSON output.
// args is a space-separated command string, e.g. "reviews list --app 123 --limit 10 --output json".
func (a *App) RunASCCommand(args string) (ASCCommandResponse, error) {
	if strings.TrimSpace(args) == "" {
		return ASCCommandResponse{Error: "args required"}, nil
	}

	ascPath, err := a.resolveASCPath()
	if err != nil {
		return ASCCommandResponse{Error: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 30*time.Second)
	defer cancel()

	parts := strings.Fields(args)
	cmd := exec.CommandContext(ctx, ascPath, parts...)
	cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ASCCommandResponse{Error: strings.TrimSpace(string(out))}, nil
	}
	return ASCCommandResponse{Data: string(out)}, nil
}

func (a *App) GetAppDetail(appID string) (AppDetail, error) {
	if strings.TrimSpace(appID) == "" {
		return AppDetail{Error: "app ID is required"}, nil
	}

	ascPath, err := a.resolveASCPath()
	if err != nil {
		return AppDetail{Error: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 30*time.Second)
	defer cancel()

	// Fetch app attrs and versions concurrently
	type attrsResult struct {
		name          string
		bundleID      string
		sku           string
		primaryLocale string
		err           error
	}
	type versionsResult struct {
		versions []AppVersion
		err      error
	}
	type subtitleRes struct {
		subtitle string
	}

	attrsCh := make(chan attrsResult, 1)
	versionsCh := make(chan versionsResult, 1)
	subtitleCh := make(chan subtitleRes, 1)

	go func() {
		cmd := exec.CommandContext(ctx, ascPath, "apps", "view", "--id", appID, "--output", "json")
		cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			attrsCh <- attrsResult{err: err}
			return
		}
		var env struct {
			Data struct {
				Attributes struct {
					Name          string `json:"name"`
					BundleID      string `json:"bundleId"`
					SKU           string `json:"sku"`
					PrimaryLocale string `json:"primaryLocale"`
				} `json:"attributes"`
			} `json:"data"`
		}
		if json.Unmarshal(out, &env) != nil {
			attrsCh <- attrsResult{err: errors.New("failed to parse app view")}
			return
		}
		a := env.Data.Attributes
		attrsCh <- attrsResult{name: a.Name, bundleID: a.BundleID, sku: a.SKU, primaryLocale: a.PrimaryLocale}
	}()

	go func() {
		cmd := exec.CommandContext(ctx, ascPath, "versions", "list", "--app", appID, "--output", "json")
		cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
		out, err := cmd.CombinedOutput()
		if err != nil {
			versionsCh <- versionsResult{err: err}
			return
		}
		type rawVersion struct {
			ID         string `json:"id"`
			Attributes struct {
				Platform        string `json:"platform"`
				VersionString   string `json:"versionString"`
				AppVersionState string `json:"appVersionState"`
				AppStoreState   string `json:"appStoreState"`
			} `json:"attributes"`
		}
		var env struct {
			Data []rawVersion `json:"data"`
		}
		if json.Unmarshal(out, &env) != nil {
			versionsCh <- versionsResult{}
			return
		}
		vs := make([]AppVersion, 0, len(env.Data))
		for _, rv := range env.Data {
			state := rv.Attributes.AppVersionState
			if state == "" {
				state = rv.Attributes.AppStoreState
			}
			vs = append(vs, AppVersion{
				ID:       rv.ID,
				Platform: rv.Attributes.Platform,
				Version:  rv.Attributes.VersionString,
				State:    state,
			})
		}
		versionsCh <- versionsResult{versions: vs}
	}()

	go func() {
		subtitleCh <- subtitleRes{subtitle: a.fetchSubtitle(ctx, ascPath, appID)}
	}()

	attrs := <-attrsCh
	vers := <-versionsCh
	sub := <-subtitleCh

	if attrs.err != nil {
		return AppDetail{Error: attrs.err.Error()}, nil
	}

	return AppDetail{
		ID:            appID,
		Name:          attrs.name,
		Subtitle:      sub.subtitle,
		BundleID:      attrs.bundleID,
		SKU:           attrs.sku,
		PrimaryLocale: attrs.primaryLocale,
		Versions:      vers.versions,
	}, nil
}

// GetVersionMetadata returns all localizations for a given App Store version.
// Pass versionID from AppVersion.ID. Returns all locales so the frontend can
// render a picker without an extra round-trip.
func (a *App) GetVersionMetadata(versionID string) (VersionMetadataResponse, error) {
	if strings.TrimSpace(versionID) == "" {
		return VersionMetadataResponse{Error: "version ID is required"}, nil
	}

	ascPath, err := a.resolveASCPath()
	if err != nil {
		return VersionMetadataResponse{Error: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ascPath, "localizations", "list",
		"--version", versionID, "--output", "json")
	cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return VersionMetadataResponse{Error: strings.TrimSpace(string(out))}, nil
	}

	type rawAttrs struct {
		Locale          string `json:"locale"`
		Description     string `json:"description"`
		Keywords        string `json:"keywords"`
		WhatsNew        string `json:"whatsNew"`
		PromotionalText string `json:"promotionalText"`
		SupportURL      string `json:"supportUrl"`
		MarketingURL    string `json:"marketingUrl"`
	}
	type rawItem struct {
		ID         string   `json:"id"`
		Attributes rawAttrs `json:"attributes"`
	}
	var envelope struct {
		Data []rawItem `json:"data"`
	}
	if json.Unmarshal(out, &envelope) != nil {
		return VersionMetadataResponse{Error: "failed to parse localizations"}, nil
	}

	locs := make([]AppLocalization, 0, len(envelope.Data))
	for _, item := range envelope.Data {
		a := item.Attributes
		locs = append(locs, AppLocalization{
			LocalizationID:  item.ID,
			Locale:          a.Locale,
			Description:     a.Description,
			Keywords:        a.Keywords,
			WhatsNew:        a.WhatsNew,
			PromotionalText: a.PromotionalText,
			SupportURL:      a.SupportURL,
			MarketingURL:    a.MarketingURL,
		})
	}
	return VersionMetadataResponse{Localizations: locs}, nil
}

// GetScreenshots returns screenshot sets for a version localization.
// Pass LocalizationID from AppLocalization.
func (a *App) GetScreenshots(localizationID string) (ScreenshotsResponse, error) {
	if strings.TrimSpace(localizationID) == "" {
		return ScreenshotsResponse{Error: "localization ID is required"}, nil
	}

	ascPath, err := a.resolveASCPath()
	if err != nil {
		return ScreenshotsResponse{Error: "Could not find asc binary: " + err.Error()}, nil
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 20*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ascPath, "screenshots", "list",
		"--version-localization", localizationID, "--output", "json")
	cmd.Env = append(os.Environ(), "ASC_BYPASS_KEYCHAIN=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return ScreenshotsResponse{Error: strings.TrimSpace(string(out))}, nil
	}

	type rawImageAsset struct {
		TemplateURL string `json:"templateUrl"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
	}
	type rawScreenshot struct {
		Attributes struct {
			ImageAsset rawImageAsset `json:"imageAsset"`
		} `json:"attributes"`
	}
	type rawSet struct {
		Set struct {
			Attributes struct {
				DisplayType string `json:"screenshotDisplayType"`
			} `json:"attributes"`
		} `json:"set"`
		Screenshots []rawScreenshot `json:"screenshots"`
	}
	var result struct {
		Sets []rawSet `json:"sets"`
	}
	if json.Unmarshal(out, &result) != nil {
		return ScreenshotsResponse{Error: "failed to parse screenshots"}, nil
	}

	sets := make([]ScreenshotSet, 0, len(result.Sets))
	for _, rs := range result.Sets {
		if len(rs.Screenshots) == 0 {
			continue
		}
		shots := make([]AppScreenshot, 0, len(rs.Screenshots))
		for _, s := range rs.Screenshots {
			ia := s.Attributes.ImageAsset
			if ia.TemplateURL == "" {
				continue
			}
			// Build a ~400px-wide thumbnail URL from the template.
			thumbW := 400
			thumbH := thumbW
			if ia.Width > 0 && ia.Height > 0 {
				thumbH = thumbW * ia.Height / ia.Width
			}
			thumbURL := strings.NewReplacer(
				"{w}", fmt.Sprintf("%d", thumbW),
				"{h}", fmt.Sprintf("%d", thumbH),
				"{f}", "webp",
			).Replace(ia.TemplateURL)
			shots = append(shots, AppScreenshot{
				ThumbnailURL: thumbURL,
				Width:        ia.Width,
				Height:       ia.Height,
			})
		}
		if len(shots) > 0 {
			sets = append(sets, ScreenshotSet{
				DisplayType: rs.Set.Attributes.DisplayType,
				Screenshots: shots,
			})
		}
	}
	return ScreenshotsResponse{Sets: sets}, nil
}

func (a *App) resolveASCPath() (string, error) {
	cfg, err := a.settings.Load()
	if err != nil {
		return "", err
	}

	bundled := filepath.Join(a.rootDir, "bin", "asc")
	resolution, err := ascbin.Resolve(ascbin.ResolveOptions{
		BundledPath:    bundled,
		SystemOverride: cfg.SystemASCPath,
		PreferBundled:  cfg.PreferBundledASC,
		LookPath:       execLookPath,
	})
	if err != nil {
		return "", err
	}
	return resolution.Path, nil
}

func (a *App) ListThreads() ([]threads.Thread, error) {
	return a.threads.LoadAll()
}

func (a *App) CreateThread(title string) (threads.Thread, error) {
	if strings.TrimSpace(title) == "" {
		title = "New Studio Thread"
	}

	now := time.Now().UTC()
	thread := threads.Thread{
		ID:        uuid.NewString(),
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := a.threads.SaveThread(thread); err != nil {
		return threads.Thread{}, err
	}
	return thread, nil
}

func (a *App) ResolveASC() (ResolutionResponse, error) {
	cfg, err := a.settings.Load()
	if err != nil {
		return ResolutionResponse{}, err
	}

	bundled := filepath.Join(a.rootDir, "bin", "asc")
	resolution, err := ascbin.Resolve(ascbin.ResolveOptions{
		BundledPath:    bundled,
		SystemOverride: cfg.SystemASCPath,
		PreferBundled:  cfg.PreferBundledASC,
		LookPath:       execLookPath,
	})
	if err != nil {
		return ResolutionResponse{}, err
	}

	return ResolutionResponse{
		Resolution:       resolution,
		AvailablePresets: settings.DefaultPresets(),
	}, nil
}

func (a *App) QueueMutation(req ApprovalRequest) (approvals.Action, error) {
	if strings.TrimSpace(req.ThreadID) == "" {
		return approvals.Action{}, errors.New("thread ID is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return approvals.Action{}, errors.New("title is required")
	}

	action := approvals.Action{
		ID:              uuid.NewString(),
		ThreadID:        req.ThreadID,
		Title:           req.Title,
		Summary:         req.Summary,
		CommandPreview:  req.CommandPreview,
		MutationSurface: req.MutationSurface,
		Status:          approvals.StatusPending,
		CreatedAt:       time.Now().UTC(),
	}
	return a.approvals.Enqueue(action), nil
}

func (a *App) ListApprovals() []approvals.Action {
	return a.approvals.Pending()
}

func (a *App) ApproveAction(id string) (approvals.Action, error) {
	return a.approvals.Approve(id)
}

func (a *App) RejectAction(id string) (approvals.Action, error) {
	return a.approvals.Reject(id)
}

func (a *App) SendPrompt(req PromptRequest) (PromptResponse, error) {
	if strings.TrimSpace(req.Prompt) == "" {
		return PromptResponse{}, errors.New("prompt is required")
	}

	thread, err := a.ensureThread(req.ThreadID)
	if err != nil {
		return PromptResponse{}, err
	}

	thread.Messages = append(thread.Messages, threads.Message{
		ID:        uuid.NewString(),
		Role:      threads.RoleUser,
		Kind:      threads.KindMessage,
		Content:   req.Prompt,
		CreatedAt: time.Now().UTC(),
	})
	thread.UpdatedAt = time.Now().UTC()
	if err := a.threads.SaveThread(thread); err != nil {
		return PromptResponse{}, err
	}

	session, err := a.ensureSession(thread)
	if err != nil {
		return PromptResponse{}, err
	}

	ctx, cancel := context.WithTimeout(a.contextOrBackground(), 15*time.Second)
	defer cancel()

	result, events, err := session.client.Prompt(ctx, session.sessionID, req.Prompt)
	if err != nil {
		return PromptResponse{}, err
	}

	for _, event := range events {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "studio:agent:update", event)
		}
	}

	assistantMessage := result.Summary()
	if assistantMessage == "" {
		assistantMessage = "ASC Studio captured the prompt and is waiting for the agent response stream."
	}

	thread.Messages = append(thread.Messages, threads.Message{
		ID:        uuid.NewString(),
		Role:      threads.RoleAssistant,
		Kind:      threads.KindMessage,
		Content:   assistantMessage,
		CreatedAt: time.Now().UTC(),
	})
	thread.SessionID = session.sessionID
	thread.UpdatedAt = time.Now().UTC()
	if err := a.threads.SaveThread(thread); err != nil {
		return PromptResponse{}, err
	}

	return PromptResponse{Thread: thread}, nil
}

func (a *App) ensureThread(id string) (threads.Thread, error) {
	if strings.TrimSpace(id) == "" {
		return a.CreateThread("New Studio Thread")
	}
	return a.threads.Get(id)
}

func (a *App) ensureSession(thread threads.Thread) (*threadSession, error) {
	a.mu.Lock()
	existing := a.sessions[thread.ID]
	a.mu.Unlock()
	if existing != nil {
		return existing, nil
	}

	cfg, err := a.settings.Load()
	if err != nil {
		return nil, err
	}
	launch, err := cfg.ResolveAgentLaunch()
	if err != nil {
		return nil, err
	}

	client, err := acp.Start(a.contextOrBackground(), acp.LaunchSpec{
		Command: launch.Command,
		Args:    launch.Args,
		Dir:     launch.Dir,
		Env:     launch.Env,
	})
	if err != nil {
		return nil, err
	}

	workspaceRoot := cfg.WorkspaceRoot
	if strings.TrimSpace(workspaceRoot) == "" {
		workspaceRoot, _ = os.Getwd()
	}

	sessionID, err := client.Bootstrap(a.contextOrBackground(), acp.SessionConfig{
		CWD:       workspaceRoot,
		SessionID: thread.SessionID,
	})
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	session := &threadSession{
		client:    client,
		sessionID: sessionID,
	}

	a.mu.Lock()
	a.sessions[thread.ID] = session
	a.mu.Unlock()

	return session, nil
}

func (a *App) contextOrBackground() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func defaultSections() []WorkspaceSection {
	return []WorkspaceSection{
		{ID: "apps", Label: "Apps", Description: "Select an app and pin its release context into Studio."},
		{ID: "overview", Label: "Overview", Description: "Monitor release readiness, metadata drift, and unresolved blockers."},
		{ID: "builds", Label: "Builds", Description: "Inspect TestFlight and App Store build status in one place."},
		{ID: "submission", Label: "Submission", Description: "Preview validation and guarded mutation flows before publish."},
		{ID: "assets", Label: "Assets", Description: "Track screenshots and localization surfaces for app store readiness."},
		{ID: "threads", Label: "Threads", Description: "Keep ACP threads, approvals, and release history together."},
	}
}

func execLookPath(file string) (string, error) {
	return execLookPathFunc(file)
}

var execLookPathFunc = func(file string) (string, error) {
	return exec.LookPath(file)
}
