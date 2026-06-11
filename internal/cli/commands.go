package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mdwcoder/core-utils-cli/internal/config"
	"github.com/mdwcoder/core-utils-cli/internal/installer"
	"github.com/mdwcoder/core-utils-cli/internal/output"
	"github.com/mdwcoder/core-utils-cli/internal/platform"
	"github.com/mdwcoder/core-utils-cli/internal/registry"
)

func Run(args []string) int {
	if len(args) == 0 {
		printHelp()
		return 0
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "help", "-h", "--help":
		printHelp()
	case "version", "-v", "--version":
		printVersion()
	case "registry":
		if len(rest) == 1 && rest[0] == "sync" {
			return cmdRegistrySync()
		}
		output.Errorln("Usage: cu registry sync")
		return 1
	case "search":
		if len(rest) < 1 {
			output.Errorln("Usage: cu search <query>")
			return 1
		}
		return cmdSearch(strings.Join(rest, " "))
	case "info":
		if len(rest) < 1 {
			output.Errorln("Usage: cu info <tool>")
			return 1
		}
		return cmdInfo(rest[0])
	case "install":
		if len(rest) < 1 {
			output.Errorln("Usage: cu install <tool>")
			return 1
		}
		return cmdInstall(rest[0])
	case "list":
		return cmdList()
	case "update":
		if len(rest) == 1 {
			return cmdUpdate(rest[0])
		}
		return cmdUpdate("")
	case "remove":
		if len(rest) < 1 {
			output.Errorln("Usage: cu remove <tool>")
			return 1
		}
		return cmdRemove(rest[0])
	case "doctor":
		return cmdDoctor()
	default:
		output.Errorf("Unknown command: %s\n", cmd)
		printHelp()
		return 1
	}
	return 0
}

func printHelp() {
	output.Println(output.Bold("Core Utils CLI (cu) — Usage:"))
	output.Println("")
	output.Println("  cu help                  Show this help")
	output.Println("  cu version               Show CLI version")
	output.Println("  cu registry sync         Download and cache the registry")
	output.Println("  cu search <query>        Search tools in the registry")
	output.Println("  cu info <tool>           Show details about a tool")
	output.Println("  cu install <tool>        Install a tool")
	output.Println("  cu list                  List installed tools")
	output.Println("  cu update [tool]         Update all or one tool")
	output.Println("  cu remove <tool>         Remove a tool")
	output.Println("  cu doctor                Diagnose the environment")
	output.Println("")
	output.Println("Environment:")
	output.Println("  CORE_UTILS_REGISTRY_URL  Override registry endpoint")
	output.Println("  NO_COLOR                 Disable colored output")
}

func printVersion() {
	output.Printf("cu version %s (%s/%s)\n", config.AppVersion, runtime.GOOS, runtime.GOARCH)
}

func cmdRegistrySync() int {
	output.Println("Fetching registry...")
	reg, err := registry.Fetch()
	if err != nil {
		output.Errorf("Failed to fetch registry: %v\n", err)
		return 1
	}
	if err := registry.SaveCache(reg); err != nil {
		output.Errorf("Failed to save cache: %v\n", err)
		return 1
	}
	output.Printf("Registry cached (%d tools).\n", len(reg.Tools))
	return 0
}

func cmdSearch(query string) int {
	reg, err := getRegistry()
	if err != nil {
		output.Errorf("%v\n", err)
		return 1
	}
	results := reg.SearchTools(query)
	if len(results) == 0 {
		output.Println("No results found.")
		return 0
	}
	output.Printf("Found %d result(s):\n", len(results))
	for _, t := range results {
		output.Printf("  %-20s %s\n", t.ID, t.Name)
	}
	return 0
}

func cmdInfo(toolID string) int {
	reg, err := getRegistry()
	if err != nil {
		output.Errorf("%v\n", err)
		return 1
	}
	t := reg.FindTool(toolID)
	if t == nil {
		output.Errorf("Tool not found: %s\n", toolID)
		return 1
	}
	output.Println(output.Bold(t.Name))
	output.Printf("  ID:          %s\n", t.ID)
	output.Printf("  Version:     %s\n", t.Version)
	output.Printf("  Description: %s\n", t.Description)
	output.Printf("  Category:    %s\n", t.Category)
	output.Printf("  Language:    %s\n", t.Language)
	output.Printf("  Binary:      %s\n", t.Binary)
	output.Printf("  Homepage:    %s\n", t.Homepage)
	output.Printf("  Repository:  %s\n", t.Repository)
	platKey := platform.Current()
	output.Printf("  Platforms:   ")
	if len(t.Platforms) == 0 {
		output.Println(" none")
	} else {
		output.Println("")
		for k := range t.Platforms {
			marker := ""
			if k == platKey {
				marker = " <- current"
			}
			output.Printf("    %s%s\n", k, marker)
		}
	}
	return 0
}

func cmdInstall(toolID string) int {
	reg, err := getRegistry()
	if err != nil {
		output.Errorf("%v\n", err)
		return 1
	}
	t := reg.FindTool(toolID)
	if t == nil {
		output.Errorf("Tool not found: %s\n", toolID)
		return 1
	}
	if err := installer.Install(t, false); err != nil {
		output.Errorf("Install failed: %v\n", err)
		return 1
	}
	return 0
}

func cmdList() int {
	if err := installer.List(); err != nil {
		output.Errorf("%v\n", err)
		return 1
	}
	return 0
}

func cmdUpdate(specific string) int {
	if err := installer.Update(specific); err != nil {
		output.Errorf("%v\n", err)
		return 1
	}
	return 0
}

func cmdRemove(toolID string) int {
	if err := installer.Remove(toolID); err != nil {
		output.Errorf("%v\n", err)
		return 1
	}
	return 0
}

func cmdDoctor() int {
	output.Println(output.Bold("cu doctor"))
	output.Println("")

	// CLI version
	output.Status("CLI version", true, fmt.Sprintf("%s (%s/%s)", config.AppVersion, runtime.GOOS, runtime.GOARCH))

	// Platform
	plat := platform.Current()
	output.Status("Platform", true, plat)

	// Directories
	config.EnsureDirs()
	dirsOK := true
	for _, d := range []string{config.BaseDir, config.BinDir, config.ToolsDir} {
		info, err := os.Stat(d)
		if err != nil {
			output.Status(fmt.Sprintf("Directory %s", d), false, err.Error())
			dirsOK = false
			continue
		}
		if !info.IsDir() {
			output.Status(fmt.Sprintf("Directory %s", d), false, "not a directory")
			dirsOK = false
			continue
		}
	}
	if dirsOK {
		output.Status("Local directories", true, config.BaseDir)
	}

	// PATH
	if installer.IsInPath(config.BinDir) {
		output.Status("PATH", true, config.BinDir+" is in PATH")
	} else {
		output.StatusWarn("PATH", config.BinDir+" is NOT in PATH")
	}

	// Registry remote
	reg, err := registry.Fetch()
	if err != nil {
		output.Status("Registry remote", false, err.Error())
	} else {
		output.Status("Registry remote", true, fmt.Sprintf("%d tools", len(reg.Tools)))
	}

	// Registry cache
	cache, err := registry.LoadCache()
	if err != nil {
		output.Status("Registry cache", false, err.Error())
	} else {
		output.Status("Registry cache", true, fmt.Sprintf("%d tools", len(cache.Tools)))
	}

	// Write permissions
	canWrite := true
	for _, d := range []string{config.BaseDir, config.BinDir, config.ToolsDir} {
		testFile := filepath.Join(d, ".write_test")
		f, err := os.Create(testFile)
		if err != nil {
			output.Status(fmt.Sprintf("Write %s", d), false, err.Error())
			canWrite = false
			continue
		}
		f.Close()
		os.Remove(testFile)
	}
	if canWrite {
		output.Status("Write permissions", true, "OK")
	}

	// Installed tools
	instDB, err := installer.LoadInstalled()
	if err != nil {
		output.Status("Installed DB", false, err.Error())
	} else {
		output.Status("Installed DB", true, fmt.Sprintf("%d tools", len(instDB.Tools)))
		for _, t := range instDB.Tools {
			binPath := filepath.Join(config.BinDir, t.Binary)
			if runtime.GOOS == "windows" && !strings.HasSuffix(binPath, ".exe") {
				binPath += ".exe"
			}
			if _, err := os.Stat(binPath); err != nil {
				output.Status(fmt.Sprintf("Binary %s", t.Binary), false, err.Error())
			} else {
				output.Status(fmt.Sprintf("Binary %s", t.Binary), true, "exists")
			}
		}
	}

	output.Println("")
	return 0
}

func getRegistry() (*registry.Registry, error) {
	reg, err := registry.Fetch()
	if err != nil {
		output.StatusWarn("Registry", "remote fetch failed, using cache")
		reg, err = registry.LoadCache()
		if err != nil {
			return nil, fmt.Errorf("cannot load registry: %w", err)
		}
	}
	return reg, nil
}
