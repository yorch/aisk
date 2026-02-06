package config

import (
	"os"
	"path/filepath"
)

// Default paths and configuration for aisk.
const (
	AppName    = "aisk"
	AppVersion = "0.1.0"
)

// Paths returns resolved paths for the application.
type Paths struct {
	Home       string // user home directory
	AiskDir    string // ~/.aisk/
	CacheDir   string // ~/.aisk/cache/
	ManifestDB string // ~/.aisk/manifest.json
	SkillsRepo string // local skills repository path
}

// ResolvePaths builds all application paths from the user's home directory.
func ResolvePaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	aiskDir := filepath.Join(home, ".aisk")
	skillsRepo := os.Getenv("AISK_SKILLS_PATH")
	if skillsRepo == "" {
		// Default: look for skills in the current working directory
		skillsRepo, err = os.Getwd()
		if err != nil {
			return Paths{}, err
		}
	}

	return Paths{
		Home:       home,
		AiskDir:    aiskDir,
		CacheDir:   filepath.Join(aiskDir, "cache"),
		ManifestDB: filepath.Join(aiskDir, "manifest.json"),
		SkillsRepo: skillsRepo,
	}, nil
}

// EnsureDirs creates the required application directories.
func (p Paths) EnsureDirs() error {
	for _, dir := range []string{p.AiskDir, p.CacheDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
