package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/audit"
	"github.com/yorch/aisk/internal/config"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Inspect aisk audit events",
	RunE:  runAudit,
}

var auditPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Prune old audit events and compact log files",
	RunE:  runAuditPrune,
}

var (
	auditLimit  int
	auditRunID  string
	auditAction string
	auditStatus string
	auditJSON   bool

	auditPruneKeepDays int
	auditPruneKeep     int
	auditPruneDryRun   bool
)

func init() {
	auditCmd.Flags().IntVar(&auditLimit, "limit", 50, "maximum number of events to show (0 = all)")
	auditCmd.Flags().StringVar(&auditRunID, "run-id", "", "filter by run ID")
	auditCmd.Flags().StringVar(&auditAction, "action", "", "filter by action")
	auditCmd.Flags().StringVar(&auditStatus, "status", "", "filter by status")
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "output as JSON")

	auditPruneCmd.Flags().IntVar(&auditPruneKeepDays, "keep-days", 30, "keep events newer than N days (0 = disable)")
	auditPruneCmd.Flags().IntVar(&auditPruneKeep, "keep", 2000, "keep at most N most recent events after filtering (0 = disable)")
	auditPruneCmd.Flags().BoolVar(&auditPruneDryRun, "dry-run", false, "preview prune results without writing")
	auditCmd.AddCommand(auditPruneCmd)
}

func runAudit(_ *cobra.Command, _ []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	logPath := resolveAuditLogPath(paths)

	events, err := loadAuditEventsWithBackups(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No audit events found.")
			return nil
		}
		return fmt.Errorf("reading audit log: %w", err)
	}

	events = filterAuditEvents(events, auditRunID, auditAction, auditStatus)
	events = tailAuditEvents(events, auditLimit)

	if len(events) == 0 {
		fmt.Println("No audit events found.")
		return nil
	}

	if auditJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(events)
	}

	return printAuditEventsTable(events)
}

func runAuditPrune(_ *cobra.Command, _ []string) error {
	if auditPruneKeepDays < 0 {
		return fmt.Errorf("--keep-days cannot be negative")
	}
	if auditPruneKeep < 0 {
		return fmt.Errorf("--keep cannot be negative")
	}

	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}
	logPath := resolveAuditLogPath(paths)

	events, err := loadAuditEventsWithBackups(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No audit events found.")
			return nil
		}
		return fmt.Errorf("reading audit log: %w", err)
	}

	originalCount := len(events)
	events = pruneByAge(events, auditPruneKeepDays)
	events = tailAuditEvents(events, auditPruneKeep)
	removed := originalCount - len(events)

	if auditPruneDryRun {
		fmt.Printf("Dry-run: would remove %d event(s), keep %d event(s).\n", removed, len(events))
		return nil
	}

	if err := writeAuditEvents(logPath, events); err != nil {
		return fmt.Errorf("writing pruned audit log: %w", err)
	}
	if err := removeAuditBackups(logPath); err != nil {
		return fmt.Errorf("removing audit backups: %w", err)
	}

	fmt.Printf("Pruned %d event(s); kept %d event(s).\n", removed, len(events))
	return nil
}

func loadAuditEventsWithBackups(primary string) ([]audit.Event, error) {
	paths := audit.CandidateLogPaths(primary)
	if len(paths) == 0 {
		return nil, os.ErrNotExist
	}

	var all []audit.Event
	for _, p := range paths {
		events, err := loadAuditEvents(p)
		if err != nil {
			return nil, err
		}
		all = append(all, events...)
	}
	return all, nil
}

func writeAuditEvents(primary string, events []audit.Event) error {
	if err := os.MkdirAll(filepath.Dir(primary), 0o755); err != nil {
		return err
	}

	f, err := os.OpenFile(primary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	for _, e := range events {
		line, err := json.Marshal(e)
		if err != nil {
			continue
		}
		if _, err := f.Write(append(line, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func removeAuditBackups(primary string) error {
	backups := audit.CandidateLogPaths(primary)
	for _, p := range backups {
		if p == primary {
			continue
		}
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func loadAuditEvents(path string) ([]audit.Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []audit.Event
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var e audit.Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func filterAuditEvents(events []audit.Event, runID, action, status string) []audit.Event {
	if runID == "" && action == "" && status == "" {
		return events
	}
	var filtered []audit.Event
	for _, e := range events {
		if runID != "" && e.RunID != runID {
			continue
		}
		if action != "" && e.Action != action {
			continue
		}
		if status != "" && e.Status != status {
			continue
		}
		filtered = append(filtered, e)
	}
	return filtered
}

func pruneByAge(events []audit.Event, keepDays int) []audit.Event {
	if keepDays <= 0 {
		return events
	}
	cutoff := time.Now().Add(-time.Duration(keepDays) * 24 * time.Hour)
	var out []audit.Event
	for _, e := range events {
		if e.Timestamp == "" {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, e.Timestamp)
		if err != nil {
			continue
		}
		if ts.After(cutoff) || ts.Equal(cutoff) {
			out = append(out, e)
		}
	}
	return out
}

func tailAuditEvents(events []audit.Event, limit int) []audit.Event {
	if limit <= 0 || len(events) <= limit {
		return events
	}
	return events[len(events)-limit:]
}

func printAuditEventsTable(events []audit.Event) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIME\tCOMMAND\tACTION\tSTATUS\tSKILL\tCLIENT")
	for _, e := range events {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			e.Timestamp,
			e.Command,
			e.Action,
			e.Status,
			e.Skill,
			e.ClientID,
		)
	}
	return w.Flush()
}

func resolveAuditLogPath(paths config.Paths) string {
	logPath := strings.TrimSpace(os.Getenv("AISK_AUDIT_LOG_PATH"))
	if logPath == "" {
		logPath = filepath.Join(paths.AiskDir, "audit.log")
	}
	return logPath
}
