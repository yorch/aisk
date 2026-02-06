package skill

import "testing"

func TestParseRepoURL(t *testing.T) {
	tests := []struct {
		input     string
		wantOwner string
		wantRepo  string
		wantOK    bool
	}{
		{"github.com/user/repo", "user", "repo", true},
		{"https://github.com/user/repo", "user", "repo", true},
		{"https://github.com/user/repo.git", "user", "repo", true},
		{"https://github.com/user/repo/", "user", "repo", true},
		{"gitlab.com/user/repo", "", "", false},
		{"not-a-url", "", "", false},
		{"github.com/user", "", "", false},
	}

	for _, tt := range tests {
		owner, repo, ok := ParseRepoURL(tt.input)
		if ok != tt.wantOK {
			t.Errorf("ParseRepoURL(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			continue
		}
		if ok {
			if owner != tt.wantOwner || repo != tt.wantRepo {
				t.Errorf("ParseRepoURL(%q) = (%q, %q), want (%q, %q)",
					tt.input, owner, repo, tt.wantOwner, tt.wantRepo)
			}
		}
	}
}
