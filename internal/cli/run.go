package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
)

// Version is set at build time with -ldflags. When QBS is installed with
// "go install ...@version", Go records the module version in the executable;
// use that value when no release-specific linker flag was supplied.
var Version = "dev"

func init() {
	if Version != "dev" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return
	}
	Version = strings.TrimPrefix(info.Main.Version, "v")
}

const usage = `qbs — AI workspace and curated skill manager

Usage:
  qbs init
  qbs provision [worktree]
  qbs skills import <path> [--force]
  qbs skills list
  qbs skills sync
  qbs skills remove <name>
  qbs task <name>
  qbs tasks
  qbs task remove <name>

Options:
  -h, --help       show this help
      --version    show version

Commands:
  update [version] download and install the latest (or selected) release
`

// Run dispatches a qbs invocation. It returns an error suitable for display
// by the executable entry point.
func Run(args []string, stdout, stderr io.Writer) error {
	return RunWithInput(args, os.Stdin, stdout, stderr)
}

// RunWithInput dispatches a qbs invocation using input for interactive prompts.
func RunWithInput(args []string, input io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		_, err := io.WriteString(stdout, usage)
		return err
	}
	if args[0] == "--version" || args[0] == "-version" {
		_, err := fmt.Fprintf(stdout, "qbs %s\n", Version)
		return err
	}
	switch args[0] {
	case "init":
		if len(args) != 1 {
			return errors.New("init does not accept arguments")
		}
		return fromCurrentDirectory(func(dir string) error {
			return initializeRepository(dir, stdout, stderr)
		})
	case "provision":
		if len(args) > 2 {
			return errors.New("provision accepts an optional worktree path")
		}
		if len(args) == 2 {
			return provisionRepository(args[1], stdout, stderr)
		}
		return fromCurrentDirectory(func(dir string) error {
			return provisionRepository(dir, stdout, stderr)
		})
	case "update":
		if len(args) > 2 {
			return errors.New("update accepts an optional version")
		}
		version := ""
		if len(args) == 2 {
			version = args[1]
		}
		return update(version, stdout)
	case "skills":
		return runSkills(args[1:], input, stdout)
	case "task":
		if len(args) >= 2 && args[1] == "remove" {
			if len(args) < 3 || len(args) > 4 || (len(args) == 4 && args[3] != "--force") {
				return errors.New("task remove requires a name and optional --force")
			}
			return fromCurrentDirectory(func(dir string) error {
				return removeTask(dir, args[2], len(args) == 4, stdout, stderr)
			})
		}
		if len(args) != 2 {
			return errors.New("task requires exactly one name")
		}
		return fromCurrentDirectory(func(dir string) error {
			return createTask(dir, args[1], stdout, stderr)
		})
	case "tasks":
		if len(args) != 1 {
			return errors.New("tasks does not accept arguments")
		}
		return fromCurrentDirectory(func(dir string) error {
			return listTasks(dir, stdout)
		})
	default:
		return fmt.Errorf("unknown command %q\n\n%s", args[0], usage)
	}
}

func runSkills(args []string, input io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("skills requires one of: import, list, sync, remove")
	}
	switch args[0] {
	case "import":
		if len(args) < 2 || len(args) > 3 || (len(args) == 3 && args[2] != "--force") {
			return errors.New("skills import requires a path and optional --force")
		}
		return importSkills(args[1], len(args) == 3, input, stdout)
	case "list":
		if len(args) != 1 {
			return errors.New("skills list does not accept arguments")
		}
		return listSkills(stdout)
	case "sync":
		if len(args) != 1 {
			return errors.New("skills sync does not accept arguments")
		}
		return syncSkills(input, stdout)
	case "remove":
		if len(args) != 2 {
			return errors.New("skills remove requires exactly one name")
		}
		return removeSkill(args[1], stdout)
	default:
		return fmt.Errorf("unknown skills command %q", args[0])
	}
}

var getwd = os.Getwd

func fromCurrentDirectory(run func(string) error) error {
	dir, err := currentDirectory()
	if err != nil {
		return err
	}
	return run(dir)
}

func currentDirectory() (string, error) {
	dir, err := getwd()
	if err != nil {
		return "", fmt.Errorf("could not determine current working directory: %w", err)
	}
	return dir, nil
}
