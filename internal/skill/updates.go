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

	// Group installations by skill name + installed version so mixed-version
	// clients are reported accurately.
	type installed struct {
		skillName        string
		installedVersion string
		availableVersion string
		clients          []string
	}
	groups := make(map[string]*installed)
	var order []string

	for _, inst := range installations {
		availVer, found := avail[inst.SkillName]
		if !found {
			continue // not in available repo
		}
		if inst.SkillVersion != "" && inst.SkillVersion != "unversioned" && inst.SkillVersion == availVer {
			continue // already up-to-date
		}

		key := inst.SkillName + "\x00" + inst.SkillVersion + "\x00" + availVer
		g, ok := groups[key]
		if !ok {
			g = &installed{
				skillName:        inst.SkillName,
				installedVersion: inst.SkillVersion,
				availableVersion: availVer,
			}
			groups[key] = g
			order = append(order, key)
		}
		g.clients = append(g.clients, inst.ClientID)
	}

	var updates []UpdateInfo
	for _, key := range order {
		g := groups[key]
		updates = append(updates, UpdateInfo{
			SkillName:        g.skillName,
			InstalledVersion: g.installedVersion,
			AvailableVersion: g.availableVersion,
			AffectedClients:  g.clients,
		})
	}

	return updates
}
