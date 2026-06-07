package version

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// repoOwner GitHub 仓库所有者，用于检测最新发布版本
	repoOwner = "NexaCard"
	// repoName GitHub 仓库名称
	repoName = "API"

	githubAPIBaseURL    = "https://api.github.com"
	githubWebBaseURL    = "https://github.com"
	releaseUserAgent    = "nexacard-update-checker"
	maxReleaseNotesRune = 20000
)

// releasePayload GitHub Releases API 响应中本检测器关心的字段
type releasePayload struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	HTMLURL     string    `json:"html_url"`
	Body        string    `json:"body"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
}

// CheckResult 检测结果，已包含当前与最新版本以及是否需要更新
type CheckResult struct {
	CurrentVersion string     `json:"current_version"`
	LatestVersion  string     `json:"latest_version"`
	HasUpdate      bool       `json:"has_update"`
	ReleaseURL     string     `json:"release_url,omitempty"`
	ReleaseName    string     `json:"release_name,omitempty"`
	ReleaseNotes   string     `json:"release_notes,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	Source         string     `json:"source"`
}

// ErrRateLimited 触发 GitHub 匿名调用速率限制时返回，便于上层映射成专用提示
var ErrRateLimited = errors.New("github api rate limit exceeded")

// CheckLatestRelease 通过 GitHub Releases API 获取最新发行版并与当前版本比较。
// 仓库地址固定为 NexaCard/API，不接受外部传入，避免 SSRF 和误触上游更新源。
func CheckLatestRelease(ctx context.Context) (*CheckResult, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", githubAPIBaseURL, repoOwner, repoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", releaseUserAgent)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request github api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return nil, ErrRateLimited
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("github release not found for %s/%s", repoOwner, repoName)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github api returned %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload releasePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}

	current := strings.TrimSpace(Version)
	latest := strings.TrimSpace(payload.TagName)
	if latest == "" {
		return nil, errors.New("github release tag is empty")
	}
	hasUpdate, _ := IsNewerVersion(latest, current)

	result := &CheckResult{
		CurrentVersion: current,
		LatestVersion:  latest,
		HasUpdate:      hasUpdate,
		ReleaseURL:     normalizeReleaseURL(payload.HTMLURL),
		ReleaseName:    payload.Name,
		ReleaseNotes:   normalizeReleaseNotes(payload.Body),
		Source:         releaseSourceURL(),
	}
	if !payload.PublishedAt.IsZero() {
		t := payload.PublishedAt
		result.PublishedAt = &t
	}
	return result, nil
}

func releaseSourceURL() string {
	return fmt.Sprintf("%s/%s/%s/releases", githubWebBaseURL, repoOwner, repoName)
}

func normalizeReleaseURL(raw string) string {
	fallback := releaseSourceURL()
	value := strings.TrimSpace(raw)
	if value == "" {
		return fallback
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed == nil {
		return fallback
	}
	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") {
		return fallback
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 3 && len(parts) != 5 {
		return fallback
	}
	if !strings.EqualFold(parts[0], repoOwner) || !strings.EqualFold(parts[1], repoName) || parts[2] != "releases" {
		return fallback
	}
	if len(parts) == 5 && (parts[3] != "tag" || strings.TrimSpace(parts[4]) == "") {
		return fallback
	}

	parsed.Scheme = "https"
	parsed.Host = "github.com"
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func normalizeReleaseNotes(raw string) string {
	notes := strings.TrimSpace(raw)
	if notes == "" {
		return ""
	}
	runes := []rune(notes)
	if len(runes) <= maxReleaseNotesRune {
		return notes
	}
	return strings.TrimSpace(string(runes[:maxReleaseNotesRune])) + "\n\n..."
}

// IsNewerVersion 判断 latest 是否比 current 更新。返回 (true, nil) 表示需要更新；
// 当任一版本号无法解析时，回退到字符串不相等比较，并返回非空 error 提示调用方。
func IsNewerVersion(latest, current string) (bool, error) {
	l, lErr := parseSemver(latest)
	c, cErr := parseSemver(current)
	if lErr != nil || cErr != nil {
		// 版本号格式无法识别时，仅在两者非空且不相等时认为有更新
		if latest == "" {
			return false, errors.New("latest version is empty")
		}
		return latest != "" && current != "" && latest != current, errors.Join(lErr, cErr)
	}

	for i := range 3 {
		if l[i] > c[i] {
			return true, nil
		}
		if l[i] < c[i] {
			return false, nil
		}
	}
	return false, nil
}

// parseSemver 将 "v1.2.3" / "1.2.3" / "v1.2.3-rc.1" 等格式解析为 [major, minor, patch]
// 仅取主.次.修订三段，忽略预发布和构建元数据
func parseSemver(v string) ([3]int, error) {
	var out [3]int
	s := strings.TrimSpace(v)
	if s == "" {
		return out, errors.New("empty version")
	}
	s = strings.TrimPrefix(s, "v")
	s = strings.TrimPrefix(s, "V")
	if i := strings.IndexAny(s, "-+"); i >= 0 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) == 0 {
		return out, fmt.Errorf("invalid version: %s", v)
	}
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(strings.TrimSpace(parts[i]))
		if err != nil {
			return out, fmt.Errorf("invalid version segment %q in %s", parts[i], v)
		}
		if n < 0 {
			return out, fmt.Errorf("negative version segment in %s", v)
		}
		out[i] = n
	}
	return out, nil
}
