package skill

import (
	"github.com/yorch/aisk/internal/manifest"
)

// UpdateInfo describes an available update for one skill.
type UpdateInfo struct {
	SkillName        string
	InstalledVersion string
	AvailableVersion string
	AffectedClients  []string
}

// CheckUpdates compares installed versions against available skills and returns mismatches.
func CheckUpdates(installations []manifest.Installation, available []*Skill) []UpdateInfo {
	// Build a map of available skill versions by name
	avail := make(map[string]string)
	for _, s := range available {
		if s.Version != "" {
			avail[s.Frontmatter.Name] = s.Version
		}
		// Also index by DirName for matching
		if s.Version != "" {
			avail[s.DirName] = s.Version
		}
	}

	// Group installations by skill name
	type installed struct {
		version string
		clients []string
	}
	groups := make(map[string]*installed)
	var order []string

	for _, inst := range installations {
		g, ok := groups[inst.SkillName]
		if !ok {
			g = &installed{version: inst.SkillVersion}
			groups[inst.SkillName] = g
			order = append(order, inst.SkillName)
		}
		g.clients = append(g.clients, inst.ClientID)
	}

	var updates []UpdateInfo
	for _, name := range order {
		g := groups[name]
		availVer, found := avail[name]
		if !found {
			continue // not in available repo
		}
		if g.version == "" || g.version == "unversioned" {
			// Installed without version — always show as updatable if repo has a version
			updates = append(updates, UpdateInfo{
				SkillName:        name,
				InstalledVersion: g.version,
				AvailableVersion: availVer,
				AffectedClients:  g.clients,
			})
			continue
		}
		if g.version != availVer {
			updates = append(updates, UpdateInfo{
				SkillName:        name,
				InstalledVersion: g.version,
				AvailableVersion: availVer,
				AffectedClients:  g.clients,
			})
		}
	}

	return updates
}
