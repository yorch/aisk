package skill

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitHubContent represents a file entry from the GitHub API.
type GitHubContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"` // "file" or "dir"
	DownloadURL string `json:"download_url"`
}

// FetchRemoteList fetches available skills from a GitHub repository.
func FetchRemoteList(owner, repo string) ([]*Skill, error) {
	client := newGitHubClient()

	// List top-level directories
	entries, err := listContents(client, owner, repo, "")
	if err != nil {
		return nil, fmt.Errorf("listing repo contents: %w", err)
	}

	var skills []*Skill
	for _, entry := range entries {
		if entry.Type != "dir" {
			continue
		}
		if strings.HasPrefix(entry.Name, ".") {
			continue
		}

		// Check for SKILL.md
		skillContent, err := fetchFile(client, owner, repo, entry.Name+"/SKILL.md")
		if err != nil {
			continue // no SKILL.md, not a skill
		}

		fm, body, err := ParseFrontmatter(skillContent)
		if err != nil {
			continue
		}

		skills = append(skills, &Skill{
			Frontmatter:  fm,
			DirName:      entry.Name,
			Source:       SourceRemote,
			MarkdownBody: body,
		})
	}

	return skills, nil
}

// FetchRemoteSkill downloads a skill from GitHub to the local cache directory.
func FetchRemoteSkill(owner, repo, cacheDir string) (*Skill, error) {
	client := newGitHubClient()

	destDir := filepath.Join(cacheDir, repo)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, err
	}

	// Download all files recursively
	if err := downloadDir(client, owner, repo, "", destDir); err != nil {
		return nil, fmt.Errorf("downloading skill: %w", err)
	}

	// Parse the downloaded SKILL.md
	skillFile := filepath.Join(destDir, "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		return nil, fmt.Errorf("reading SKILL.md: %w", err)
	}

	fm, body, err := ParseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing SKILL.md: %w", err)
	}

	s := &Skill{
		Frontmatter:  fm,
		DirName:      repo,
		Path:         destDir,
		Source:       SourceRemote,
		MarkdownBody: body,
	}

	s.ReferenceFiles = discoverFiles(destDir, "reference")
	if len(s.ReferenceFiles) == 0 {
		s.ReferenceFiles = discoverFiles(destDir, "references")
	}
	s.ExampleFiles = discoverFiles(destDir, "examples")
	s.AssetFiles = discoverFiles(destDir, "assets")

	return s, nil
}

// ParseRepoURL extracts owner/repo from "github.com/owner/repo" format.
func ParseRepoURL(url string) (owner, repo string, ok bool) {
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")

	parts := strings.Split(url, "/")
	if len(parts) < 3 || parts[0] != "github.com" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func newGitHubClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func githubHeaders() http.Header {
	h := http.Header{}
	h.Set("Accept", "application/vnd.github.v3+json")
	h.Set("User-Agent", "aisk/0.1.0")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		h.Set("Authorization", "Bearer "+token)
	}
	return h
}

func listContents(client *http.Client, owner, repo, path string) ([]GitHubContent, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = githubHeaders()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var entries []GitHubContent
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func fetchFile(client *http.Client, owner, repo, path string) (string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "aisk/0.1.0")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d for %s", resp.StatusCode, path)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func downloadDir(client *http.Client, owner, repo, path, destDir string) error {
	entries, err := listContents(client, owner, repo, path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Type == "dir" {
			subDir := filepath.Join(destDir, entry.Name)
			if err := os.MkdirAll(subDir, 0o755); err != nil {
				return err
			}
			if err := downloadDir(client, owner, repo, entry.Path, subDir); err != nil {
				return err
			}
		} else if entry.DownloadURL != "" {
			content, err := downloadURL(client, entry.DownloadURL)
			if err != nil {
				return fmt.Errorf("downloading %s: %w", entry.Path, err)
			}
			dest := filepath.Join(destDir, entry.Name)
			if err := os.WriteFile(dest, content, 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadURL(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "aisk/0.1.0")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
