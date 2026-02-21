package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/audit"
	"github.com/yorch/aisk/internal/config"
)

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Inspect aisk audit events",
	RunE:  runAudit,
}

var (
	auditLimit  int
	auditRunID  string
	auditAction string
	auditStatus string
	auditJSON   bool
)

func init() {
	auditCmd.Flags().IntVar(&auditLimit, "limit", 50, "maximum number of events to show (0 = all)")
	auditCmd.Flags().StringVar(&auditRunID, "run-id", "", "filter by run ID")
	auditCmd.Flags().StringVar(&auditAction, "action", "", "filter by action")
	auditCmd.Flags().StringVar(&auditStatus, "status", "", "filter by status")
	auditCmd.Flags().BoolVar(&auditJSON, "json", false, "output as JSON")
}

func runAudit(_ *cobra.Command, _ []string) error {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	logPath := strings.TrimSpace(os.Getenv("AISK_AUDIT_LOG_PATH"))
	if logPath == "" {
		logPath = filepath.Join(paths.AiskDir, "audit.log")
	}

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
