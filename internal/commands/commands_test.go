package commands

import (
	"strings"
	"testing"
)

func TestCmdDetectBasic(t *testing.T) {
	names, shortHelp, longHelp, execute := CmdDetect()
	
	if len(names) == 0 {
		t.Error("CmdDetect should return at least one name")
	}
	
	if shortHelp == "" {
		t.Error("CmdDetect should have short help")
	}
	
	if execute == nil {
		t.Error("CmdDetect should have execute function")
	}
	
	// Basic validation that we got some help text
	_ = longHelp // Use the variable to avoid unused error
}

func TestCmdBuildBasic(t *testing.T) {
	names, _, longHelp, execute := CmdBuild()
	
	if len(names) == 0 {
		t.Error("CmdBuild should return at least one name")
	}
	
	if execute == nil {
		t.Error("CmdBuild should have execute function")
	}
	
	// Check for key build targets in help
	if !strings.Contains(longHelp, "curseforge") {
		t.Error("Build help should mention curseforge")
	}
}

func TestCmdVersionBasic(t *testing.T) {
	names, shortHelp, _, execute := CmdVersion()
	
	if len(names) == 0 {
		t.Error("CmdVersion should return at least one name")
	}
	
	if shortHelp == "" {
		t.Error("CmdVersion should have short help")
	}
	
	if execute == nil {
		t.Error("CmdVersion should have execute function")
	}
}

func TestCmdHelpBasic(t *testing.T) {
	names, _, _, execute := CmdHelp()
	
	if len(names) == 0 {
		t.Error("CmdHelp should return at least one name")
	}
	
	if execute == nil {
		t.Error("CmdHelp should have execute function")
	}
}

func TestCmdModBasic(t *testing.T) {
	names, _, _, execute := CmdMod()
	
	if len(names) == 0 {
		t.Error("CmdMod should return at least one name")
	}
	
	if execute == nil {
		t.Error("CmdMod should have execute function")
	}
}
