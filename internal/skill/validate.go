package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Severity indicates the severity level of a lint finding.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
)

func (s Severity) String() string {
	if s == SeverityError {
		return "error"
	}
	return "warning"
}

// LintResult is a single finding from linting.
type LintResult struct {
	Severity Severity
	Field    string
	Message  string
}

// LintReport collects all lint findings.
type LintReport struct {
	Results []LintResult
}

// HasErrors returns true if the report contains any errors.
func (r *LintReport) HasErrors() bool {
	for _, res := range r.Results {
		if res.Severity == SeverityError {
			return true
		}
	}
	return false
}

// Errors returns only error-level results.
func (r *LintReport) Errors() []LintResult {
	var errs []LintResult
	for _, res := range r.Results {
		if res.Severity == SeverityError {
			errs = append(errs, res)
		}
	}
	return errs
}

// Warnings returns only warning-level results.
func (r *LintReport) Warnings() []LintResult {
	var warns []LintResult
	for _, res := range r.Results {
		if res.Severity == SeverityWarning {
			warns = append(warns, res)
		}
	}
	return warns
}

var nameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

const maxNameLen = 64

// ValidateName checks that a skill name is valid kebab-case.
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) > maxNameLen {
		return fmt.Errorf("name exceeds %d characters", maxNameLen)
	}
	if !nameRegex.MatchString(name) {
		return fmt.Errorf("name must be kebab-case (lowercase letters, digits, hyphens): %q", name)
	}
	return nil
}

var semverLoose = regexp.MustCompile(`^\d+\.\d+\.\d+`)

// LintSkillMD validates the content of a SKILL.md file.
func LintSkillMD(content string) *LintReport {
	r := &LintReport{}

	fm, body, err := ParseFrontmatter(content)
	if err != nil {
		r.Results = append(r.Results, LintResult{
			Severity: SeverityError,
			Field:    "frontmatter",
			Message:  fmt.Sprintf("invalid frontmatter: %v", err),
		})
		return r
	}

	// Validate name
	if err := ValidateName(fm.Name); err != nil {
		r.Results = append(r.Results, LintResult{
			Severity: SeverityError,
			Field:    "name",
			Message:  err.Error(),
		})
	}

	// Validate description
	if strings.TrimSpace(fm.Description) == "" {
		r.Results = append(r.Results, LintResult{
			Severity: SeverityError,
			Field:    "description",
			Message:  "description is required",
		})
	}

	// Validate version format
	if fm.Version != "" && !semverLoose.MatchString(fm.Version) {
		r.Results = append(r.Results, LintResult{
			Severity: SeverityWarning,
			Field:    "version",
			Message:  fmt.Sprintf("version %q does not follow semver format (X.Y.Z)", fm.Version),
		})
	}

	// Validate body is not empty
	if strings.TrimSpace(body) == "" {
		r.Results = append(r.Results, LintResult{
			Severity: SeverityError,
			Field:    "body",
			Message:  "SKILL.md body is empty",
		})
	}

	// Warn if no "Use when:" trigger section
	if !strings.Contains(body, "Use when:") && !strings.Contains(body, "use when:") {
		r.Results = append(r.Results, LintResult{
			Severity: SeverityWarning,
			Field:    "body",
			Message:  "no \"Use when:\" trigger section found",
		})
	}

	return r
}

// LintSkillDir validates a skill directory.
func LintSkillDir(dirPath string) (*LintReport, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access %s: %w", dirPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dirPath)
	}

	skillFile := filepath.Join(dirPath, "SKILL.md")
	data, err := os.ReadFile(skillFile)
	if err != nil {
		r := &LintReport{}
		r.Results = append(r.Results, LintResult{
			Severity: SeverityError,
			Field:    "SKILL.md",
			Message:  "SKILL.md not found in directory",
		})
		return r, nil
	}

	report := LintSkillMD(string(data))

	// Warn on empty reference/ directory
	refDir := filepath.Join(dirPath, "reference")
	if isEmpty, _ := isDirEmpty(refDir); isEmpty {
		report.Results = append(report.Results, LintResult{
			Severity: SeverityWarning,
			Field:    "reference/",
			Message:  "reference directory is empty",
		})
	}

	// Warn on empty examples/ directory
	exDir := filepath.Join(dirPath, "examples")
	if isEmpty, _ := isDirEmpty(exDir); isEmpty {
		report.Results = append(report.Results, LintResult{
			Severity: SeverityWarning,
			Field:    "examples/",
			Message:  "examples directory is empty",
		})
	}

	return report, nil
}

// isDirEmpty returns true if the directory exists and is empty.
func isDirEmpty(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false, err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
