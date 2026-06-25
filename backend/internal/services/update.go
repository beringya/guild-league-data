package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nsh-guild-analytics/backend/internal/config"
)

type UpdateInfo struct {
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	Channel         string    `json:"channel"`
	ReleaseURL      string    `json:"release_url,omitempty"`
	DownloadURL     string    `json:"download_url,omitempty"`
	Checksum        string    `json:"checksum,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	InstallCommand  string    `json:"install_command,omitempty"`
	Source          string    `json:"source"`
	CheckedAt       time.Time `json:"checked_at"`
	Error           string    `json:"error,omitempty"`
}

type updateManifest struct {
	Version       string        `json:"version"`
	LatestVersion string        `json:"latest_version"`
	Channel       string        `json:"channel"`
	ReleaseURL    string        `json:"release_url"`
	DownloadURL   string        `json:"download_url"`
	URL           string        `json:"url"`
	Checksum      string        `json:"checksum"`
	Notes         string        `json:"notes"`
	Assets        []updateAsset `json:"assets"`
}

type updateAsset struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	DownloadURL string `json:"download_url"`
}

type githubRelease struct {
	TagName    string        `json:"tag_name"`
	HTMLURL    string        `json:"html_url"`
	TarballURL string        `json:"tarball_url"`
	ZipballURL string        `json:"zipball_url"`
	Body       string        `json:"body"`
	Assets     []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func CheckUpdate(ctx context.Context, cfg config.Config) UpdateInfo {
	info := UpdateInfo{
		CurrentVersion: cfg.AppVersion,
		LatestVersion:  cfg.AppVersion,
		Channel:        cfg.UpdateChannel,
		InstallCommand: cfg.UpdateInstallCommand,
		Source:         "local",
		CheckedAt:      time.Now().UTC(),
	}
	if cfg.UpdateCheckURL == "" && cfg.UpdateGithubRepo == "" {
		info.Error = "update_check_not_configured"
		return info
	}

	timeout := cfg.UpdateCheckTimeout
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	if cfg.UpdateGithubRepo != "" {
		manifest, err := fetchGithubRelease(ctx, client, cfg)
		if err != nil {
			info.Error = err.Error()
			return info
		}
		applyManifest(&info, manifest, cfg, "github")
		return info
	}

	manifest, err := fetchUpdateManifest(ctx, client, cfg)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	applyManifest(&info, manifest, cfg, "manifest")
	return info
}

func fetchUpdateManifest(ctx context.Context, client *http.Client, cfg config.Config) (updateManifest, error) {
	var manifest updateManifest
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UpdateCheckURL, nil)
	if err != nil {
		return manifest, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "nsh-guild-analytics/"+cfg.AppVersion)
	resp, err := client.Do(req)
	if err != nil {
		return manifest, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return manifest, fmt.Errorf("update check failed: HTTP %d", resp.StatusCode)
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func fetchGithubRelease(ctx context.Context, client *http.Client, cfg config.Config) (updateManifest, error) {
	parts := strings.Split(strings.Trim(strings.TrimSpace(cfg.UpdateGithubRepo), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return updateManifest{}, fmt.Errorf("UPDATE_GITHUB_REPO must use owner/repo")
	}
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", url.PathEscape(parts[0]), url.PathEscape(parts[1]))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return updateManifest{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "nsh-guild-analytics/"+cfg.AppVersion)
	resp, err := client.Do(req)
	if err != nil {
		return updateManifest{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return updateManifest{}, fmt.Errorf("github release check failed: HTTP %d", resp.StatusCode)
	}
	var release githubRelease
	if err = json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return updateManifest{}, err
	}
	assets := make([]updateAsset, 0, len(release.Assets))
	for _, asset := range release.Assets {
		assets = append(assets, updateAsset{Name: asset.Name, DownloadURL: asset.BrowserDownloadURL})
	}
	return updateManifest{
		Version:     release.TagName,
		ReleaseURL:  release.HTMLURL,
		DownloadURL: firstNonEmpty(bestAssetDownload(assets), release.TarballURL, release.ZipballURL),
		Notes:       release.Body,
		Assets:      assets,
	}, nil
}

func applyManifest(info *UpdateInfo, manifest updateManifest, cfg config.Config, source string) {
	latest := firstNonEmpty(manifest.LatestVersion, manifest.Version)
	if latest == "" {
		info.Error = "latest_version_missing"
		return
	}
	info.LatestVersion = latest
	info.Channel = firstNonEmpty(manifest.Channel, cfg.UpdateChannel)
	info.ReleaseURL = firstNonEmpty(manifest.ReleaseURL, manifest.URL)
	info.DownloadURL = firstNonEmpty(manifest.DownloadURL, cfg.UpdateDownloadURL, bestAssetDownload(manifest.Assets), info.ReleaseURL)
	info.Checksum = manifest.Checksum
	info.Notes = manifest.Notes
	info.Source = source
	info.UpdateAvailable = compareVersions(latest, cfg.AppVersion) > 0
}

func bestAssetDownload(assets []updateAsset) string {
	preferred := []string{".tar.gz", ".tgz", ".zip", ".7z"}
	for _, suffix := range preferred {
		for _, asset := range assets {
			name := strings.ToLower(asset.Name)
			if strings.HasSuffix(name, suffix) {
				return firstNonEmpty(asset.DownloadURL, asset.URL)
			}
		}
	}
	for _, asset := range assets {
		if download := firstNonEmpty(asset.DownloadURL, asset.URL); download != "" {
			return download
		}
	}
	return ""
}

func compareVersions(left, right string) int {
	a := versionSegments(left)
	b := versionSegments(right)
	size := len(a)
	if len(b) > size {
		size = len(b)
	}
	for i := 0; i < size; i++ {
		av, bv := 0, 0
		if i < len(a) {
			av = a[i]
		}
		if i < len(b) {
			bv = b[i]
		}
		if av > bv {
			return 1
		}
		if av < bv {
			return -1
		}
	}
	return 0
}

func versionSegments(raw string) []int {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "v")
	value = strings.TrimPrefix(value, "V")
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == '+'
	})
	out := make([]int, 0, len(fields))
	for _, field := range fields {
		digits := leadingDigits(field)
		if digits == "" {
			continue
		}
		value, err := strconv.Atoi(digits)
		if err == nil {
			out = append(out, value)
		}
	}
	if len(out) == 0 {
		return []int{0}
	}
	return out
}

func leadingDigits(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if r < '0' || r > '9' {
			break
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
