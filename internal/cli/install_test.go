package cli

import (
	"strings"
	"testing"
)

func TestValidateInstallNonInteractive_RequiresSkillArg(t *testing.T) {
	origAssumeYes, origInstallClient := assumeYes, installClient
	t.Cleanup(func() {
		assumeYes, installClient = origAssumeYes, origInstallClient
	})

	assumeYes = true
	installClient = "claude"

	err := validateInstallNonInteractive(nil)
	if err == nil || !strings.Contains(err.Error(), "skill argument is required") {
		t.Fatalf("expected skill argument validation error, got: %v", err)
	}
}

func TestValidateInstallNonInteractive_RequiresClient(t *testing.T) {
	origAssumeYes, origInstallClient := assumeYes, installClient
	t.Cleanup(func() {
		assumeYes, installClient = origAssumeYes, origInstallClient
	})

	assumeYes = true
	installClient = ""

	err := validateInstallNonInteractive([]string{"skill-a"})
	if err == nil || !strings.Contains(err.Error(), "--client is required") {
		t.Fatalf("expected client validation error, got: %v", err)
	}
}

func TestValidateInstallNonInteractive_AllowsExplicitInputs(t *testing.T) {
	origAssumeYes, origInstallClient := assumeYes, installClient
	t.Cleanup(func() {
		assumeYes, installClient = origAssumeYes, origInstallClient
	})

	assumeYes = true
	installClient = "claude"

	if err := validateInstallNonInteractive([]string{"skill-a"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
