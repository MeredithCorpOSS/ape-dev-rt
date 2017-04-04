package main

import (
	"testing"
)

func TestRemove(t *testing.T) {
	exitCode := runCommand([]string{"arukas", "rm", "2b21fe34-328f-4d7e-8678-726d9eff2b7f"})
	if exitCode != 0 {
		t.Errorf(("ExitCode got: %d, want: 0"), exitCode)
	}
}
