package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yorch/aisk/internal/adapter"
	"github.com/yorch/aisk/internal/audit"
	"github.com/yorch/aisk/internal/client"
	"github.com/yorch/aisk/internal/config"
	"github.com/yorch/aisk/internal/manifest"
	"github.com/yorch/aisk/internal/skill"
	"github.com/yorch/aisk/internal/tui"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Preview planned install/update/uninstall changes without writing",
}

var planInstallCmd = &cobra.Command{
	Use:   "install [skill]",
	Short: "Preview install changes",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runPlanInstall,
}

var planUpdateCmd = &cobra.Command{
	Use:   "update [skill]",
	Short: "Preview update changes",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runPlanUpdate,
}

var planUninstallCmd = &cobra.Command{
	Use:   "uninstall <skill>",
	Short: "Preview uninstall changes",
	Args:  cobra.ExactArgs(1),
	RunE:  runPlanUninstall,
}

var (
	planInstallClient      string
	planInstallScope       string
	planInstallIncludeRefs bool
	planUpdateClient       string
	planUninstallClient    string
)

func init() {
	planInstallCmd.Flags().StringVar(&planInstallClient, "client", "", "target client (claude, gemini, codex, copilot, cursor, windsurf)")
	planInstallCmd.Flags().StringVar(&planInstallScope, "scope", "global", "installation scope (global or project)")
	planInstallCmd.Flags().BoolVar(&planInstallIncludeRefs, "include-refs", false, "inline reference files in output")

	planUpdateCmd.Flags().StringVar(&planUpdateClient, "client", "", "specific client to update")
	planUninstallCmd.Flags().StringVar(&planUninstallClient, "client", "", "specific client to uninstall from")

	planCmd.AddCommand(planInstallCmd)
	planCmd.AddCommand(planUpdateCmd)
	planCmd.AddCommand(planUninstallCmd)
}

func runPlanInstall(_ *cobra.Command, args []string) (retErr error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	al := audit.New(paths.AiskDir, "plan")
	al.Log("command.plan", "started", map[string]any{
		"mode":         "install",
		"args":         args,
		"client":       planInstallClient,
		"scope":        planInstallScope,
		"include_refs": planInstallIncludeRefs,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.plan", status, map[string]any{"mode": "install"}, retErr)
	}()

	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}
	if len(skills) == 0 {
		return fmt.Errorf("no skills found in %s", paths.SkillsRepo)
	}

	target, err := resolvePlanSkill(args, skills)
	if err != nil {
		return err
	}

	reg := client.NewRegistry()
	client.DetectAll(reg, paths.Home)

	targetClients, err := resolvePlanInstallClients(reg, target.Frontmatter.Name)
	if err != nil {
		return err
	}

	fmt.Printf("Plan (install): %q\n", target.Frontmatter.Name)
	fmt.Printf("Scope: %s\n", planInstallScope)
	for _, c := range targetClients {
		targetPath := resolveTargetPath(c, planInstallScope)
		if targetPath == "" {
			fmt.Printf("- %s (%s): skipped (does not support %s scope)\n", c.Name, c.ID, planInstallScope)
			continue
		}

		adp, err := adapter.ForClient(c.ID)
		if err != nil {
			fmt.Printf("- %s (%s): error (%v)\n", c.Name, c.ID, err)
			continue
		}

		opts := adapter.InstallOpts{
			Scope:       planInstallScope,
			IncludeRefs: planInstallIncludeRefs,
			DryRun:      true,
		}
		desc := adp.Describe(target, targetPath, opts)
		op := inferInstallOperation(c.ID, targetPath, target, planInstallScope)
		fmt.Printf("- %s (%s): %s\n", c.Name, c.ID, op)
		fmt.Printf("  adapter: %s\n", desc)
	}

	return nil
}

func runPlanUpdate(_ *cobra.Command, args []string) (retErr error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	al := audit.New(paths.AiskDir, "plan")
	al.Log("command.plan", "started", map[string]any{
		"mode":   "update",
		"args":   args,
		"client": planUpdateClient,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.plan", status, map[string]any{"mode": "update"}, retErr)
	}()

	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	skills, err := skill.ScanLocal(paths.SkillsRepo)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	skillMap := make(map[string]*skill.Skill)
	for _, s := range skills {
		skillMap[s.Frontmatter.Name] = s
		skillMap[s.DirName] = s
	}

	var targets []manifest.Installation
	if len(args) > 0 {
		targets = m.Find(args[0], planUpdateClient)
		if len(targets) == 0 {
			if s, ok := skillMap[args[0]]; ok {
				targets = m.Find(s.Frontmatter.Name, planUpdateClient)
			}
		}
	} else {
		targets = m.Installations
		if planUpdateClient != "" {
			targets = m.FindByClient(planUpdateClient)
		}
	}

	if len(targets) == 0 {
		fmt.Println("No matching installations to update.")
		return nil
	}

	fmt.Println("Plan (update):")
	for _, inst := range targets {
		s := skillMap[inst.SkillName]
		if s == nil {
			fmt.Printf("- %s on %s: skipped (skill not found in local repo)\n", inst.SkillName, inst.ClientID)
			continue
		}

		clientID := client.ParseClientID(inst.ClientID)
		adp, err := adapter.ForClient(clientID)
		if err != nil {
			fmt.Printf("- %s on %s: error (%v)\n", inst.SkillName, inst.ClientID, err)
			continue
		}

		opts := adapter.InstallOpts{Scope: inst.Scope}
		desc := adp.Describe(s, inst.InstallPath, opts)
		op := inferInstallOperation(clientID, inst.InstallPath, s, inst.Scope)
		versionNote := "no version change"
		if inst.SkillVersion != s.DisplayVersion() {
			versionNote = fmt.Sprintf("%s -> %s", inst.SkillVersion, s.DisplayVersion())
		}

		fmt.Printf("- %s on %s (%s): %s [%s]\n", inst.SkillName, inst.ClientID, inst.Scope, op, versionNote)
		fmt.Printf("  adapter: %s\n", desc)
	}

	return nil
}

func runPlanUninstall(_ *cobra.Command, args []string) (retErr error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return err
	}

	al := audit.New(paths.AiskDir, "plan")
	al.Log("command.plan", "started", map[string]any{
		"mode":   "uninstall",
		"args":   args,
		"client": planUninstallClient,
	}, nil)
	defer func() {
		status := "success"
		if retErr != nil {
			status = "error"
		}
		al.Log("command.plan", status, map[string]any{"mode": "uninstall"}, retErr)
	}()

	skillArg := args[0]
	m, err := manifest.Load(paths.ManifestDB)
	if err != nil {
		return fmt.Errorf("loading manifest: %w", err)
	}

	installations := m.Find(skillArg, planUninstallClient)
	skillMap, err := scanSkillsByName(paths.SkillsRepo)
	if err != nil {
		return fmt.Errorf("scanning skills: %w", err)
	}

	if len(installations) == 0 {
		for _, s := range skillMap {
			if s.DirName == skillArg {
				installations = m.Find(s.Frontmatter.Name, planUninstallClient)
				skillArg = s.Frontmatter.Name
				break
			}
		}
	}

	if len(installations) == 0 {
		return fmt.Errorf("no installations found for %q", skillArg)
	}

	fmt.Printf("Plan (uninstall): %q\n", skillArg)
	for _, inst := range installations {
		s := skillMap[inst.SkillName]
		op := inferUninstallOperation(inst, s)
		fmt.Printf("- %s (%s): %s\n", inst.ClientID, inst.Scope, op)
	}

	return nil
}

func resolvePlanSkill(args []string, skills []*skill.Skill) (*skill.Skill, error) {
	if len(args) == 0 {
		if assumeYes {
			return nil, fmt.Errorf("skill argument is required when --yes is set")
		}
		return tui.RunSkillSelect(skills)
	}

	skillArg := args[0]
	for _, s := range skills {
		if s.DirName == skillArg || s.Frontmatter.Name == skillArg {
			return s, nil
		}
	}
	return nil, fmt.Errorf("skill %q not found", skillArg)
}

func resolvePlanInstallClients(reg *client.Registry, skillName string) ([]*client.Client, error) {
	if planInstallClient == "" {
		if assumeYes {
			return nil, fmt.Errorf("--client is required when --yes is set")
		}
		detected := reg.Detected()
		if len(detected) == 0 {
			return nil, fmt.Errorf("no AI clients detected on this system")
		}
		title := fmt.Sprintf("Plan install %q to:", skillName)
		selected, err := tui.RunClientSelect(title, detected)
		if err != nil {
			return nil, err
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf("no clients selected")
		}
		return selected, nil
	}

	clientID := client.ParseClientID(planInstallClient)
	if clientID == "" {
		return nil, fmt.Errorf("unknown client %q (valid: claude, gemini, codex, copilot, cursor, windsurf)", planInstallClient)
	}
	c := reg.Get(clientID)
	if !c.Detected {
		return nil, fmt.Errorf("client %s not detected on this system", c.Name)
	}
	return []*client.Client{c}, nil
}

func inferInstallOperation(clientID client.ClientID, targetPath string, s *skill.Skill, scope string) string {
	if isSectionBasedClient(clientID, scope) {
		switch inferSectionInstallOperation(targetPath, s.Frontmatter.Name) {
		case "create":
			return fmt.Sprintf("create %s with managed section", targetPath)
		case "replace":
			return fmt.Sprintf("replace existing managed section in %s", targetPath)
		default:
			return fmt.Sprintf("append managed section to %s", targetPath)
		}
	}

	switch clientID {
	case client.Claude:
		dest := filepath.Join(targetPath, s.DirName)
		if pathExists(dest) {
			return fmt.Sprintf("replace existing skill directory %s", dest)
		}
		return fmt.Sprintf("create skill directory %s", dest)
	case client.Cursor:
		dest := filepath.Join(targetPath, s.DirName+".mdc")
		if pathExists(dest) {
			return fmt.Sprintf("replace existing rule file %s", dest)
		}
		return fmt.Sprintf("create rule file %s", dest)
	case client.Windsurf:
		dest := filepath.Join(targetPath, s.DirName+".md")
		if pathExists(dest) {
			return fmt.Sprintf("replace existing rule file %s", dest)
		}
		return fmt.Sprintf("create rule file %s", dest)
	default:
		return fmt.Sprintf("apply install to %s", targetPath)
	}
}

func inferSectionInstallOperation(targetPath, skillName string) string {
	data, err := os.ReadFile(targetPath)
	if os.IsNotExist(err) {
		return "create"
	}
	if err != nil {
		return "append"
	}

	startMarker := fmt.Sprintf("<!-- aisk:start:%s -->", skillName)
	endMarker := fmt.Sprintf("<!-- aisk:end:%s -->", skillName)
	content := string(data)
	if strings.Contains(content, startMarker) && strings.Contains(content, endMarker) {
		return "replace"
	}
	return "append"
}

func inferUninstallOperation(inst manifest.Installation, s *skill.Skill) string {
	clientID := client.ParseClientID(inst.ClientID)
	dirName := installationDirName(s, inst.SkillName)

	switch clientID {
	case client.Claude:
		return fmt.Sprintf("remove directory %s", filepath.Join(inst.InstallPath, dirName))
	case client.Gemini, client.Codex, client.Copilot:
		return fmt.Sprintf("remove managed section from %s", inst.InstallPath)
	case client.Cursor:
		return fmt.Sprintf("remove file %s", filepath.Join(inst.InstallPath, dirName+".mdc"))
	case client.Windsurf:
		if inst.Scope == "global" {
			return fmt.Sprintf("remove managed section from %s", inst.InstallPath)
		}
		return fmt.Sprintf("remove file %s", filepath.Join(inst.InstallPath, dirName+".md"))
	default:
		return fmt.Sprintf("remove installation at %s", inst.InstallPath)
	}
}

func scanSkillsByName(skillsRepo string) (map[string]*skill.Skill, error) {
	skills, err := skill.ScanLocal(skillsRepo)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*skill.Skill, len(skills))
	for _, s := range skills {
		result[s.Frontmatter.Name] = s
	}
	return result, nil
}

func installationDirName(s *skill.Skill, fallback string) string {
	if s != nil && s.DirName != "" {
		return s.DirName
	}
	f := strings.TrimSpace(strings.ToLower(fallback))
	f = strings.ReplaceAll(f, " ", "-")
	if f == "" {
		return "skill"
	}
	return f
}

func isSectionBasedClient(clientID client.ClientID, scope string) bool {
	if clientID == client.Windsurf {
		return scope == "global"
	}
	return clientID == client.Gemini || clientID == client.Codex || clientID == client.Copilot
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
