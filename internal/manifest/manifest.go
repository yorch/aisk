package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Installation tracks a single skill installation.
type Installation struct {
	SkillName    string    `json:"skill_name"`
	SkillVersion string    `json:"skill_version"`
	ClientID     string    `json:"client_id"`
	Scope        string    `json:"scope"`
	InstalledAt  time.Time `json:"installed_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	InstallPath  string    `json:"install_path"`
}

// Manifest holds all tracked installations.
type Manifest struct {
	Installations []Installation `json:"installations"`
	path          string
}

// Load reads the manifest from disk, or returns an empty manifest.
func Load(path string) (*Manifest, error) {
	m := &Manifest{path: path}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return m, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}

	return m, nil
}

// Save writes the manifest to disk.
func (m *Manifest) Save() error {
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0o644)
}

// Add records a new installation, replacing any existing entry for the same skill+client+scope.
func (m *Manifest) Add(inst Installation) {
	m.Remove(inst.SkillName, inst.ClientID, inst.Scope)
	m.Installations = append(m.Installations, inst)
}

// Remove deletes an installation entry.
func (m *Manifest) Remove(skillName, clientID, scope string) {
	filtered := m.Installations[:0]
	for _, inst := range m.Installations {
		if inst.SkillName == skillName && inst.ClientID == clientID && inst.Scope == scope {
			continue
		}
		filtered = append(filtered, inst)
	}
	m.Installations = filtered
}

// RemoveAll deletes all installation entries for a skill.
func (m *Manifest) RemoveAll(skillName string) {
	filtered := m.Installations[:0]
	for _, inst := range m.Installations {
		if inst.SkillName == skillName {
			continue
		}
		filtered = append(filtered, inst)
	}
	m.Installations = filtered
}

// Find returns installations matching the given skill name, optionally filtered by client.
func (m *Manifest) Find(skillName string, clientID string) []Installation {
	var result []Installation
	for _, inst := range m.Installations {
		if inst.SkillName != skillName {
			continue
		}
		if clientID != "" && inst.ClientID != clientID {
			continue
		}
		result = append(result, inst)
	}
	return result
}

// FindByClient returns all installations for a given client.
func (m *Manifest) FindByClient(clientID string) []Installation {
	var result []Installation
	for _, inst := range m.Installations {
		if inst.ClientID == clientID {
			result = append(result, inst)
		}
	}
	return result
}

// AllSkillNames returns a deduplicated list of installed skill names.
func (m *Manifest) AllSkillNames() []string {
	seen := make(map[string]bool)
	var names []string
	for _, inst := range m.Installations {
		if !seen[inst.SkillName] {
			seen[inst.SkillName] = true
			names = append(names, inst.SkillName)
		}
	}
	return names
}
