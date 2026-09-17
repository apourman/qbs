package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/trues/qbs/internal/skills"
)

func importSkills(source string, force bool, input io.Reader, stdout io.Writer) error {
	catalog, err := skills.DefaultCatalog()
	if err != nil {
		return err
	}
	names, err := catalog.ImportInteractive(source, force, replacementPrompt(input, stdout))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Imported and synchronized %d skill(s): %s\n", len(names), joinNames(names))
	return err
}

func listSkills(stdout io.Writer) error {
	catalog, err := skills.DefaultCatalog()
	if err != nil {
		return err
	}
	names, err := catalog.List()
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(stdout, "Catalog: %s\n", catalog.Root); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, "Global targets:"); err != nil {
		return err
	}
	for _, target := range catalog.Targets {
		if _, err := fmt.Fprintf(stdout, "  %s\n", target); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(stdout, "Skills:"); err != nil {
		return err
	}
	for _, name := range names {
		if _, err := fmt.Fprintf(stdout, "  %s\n", name); err != nil {
			return err
		}
	}
	return nil
}

func syncSkills(input io.Reader, stdout io.Writer) error {
	catalog, err := skills.DefaultCatalog()
	if err != nil {
		return err
	}
	names, err := catalog.SyncInteractive(replacementPrompt(input, stdout))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Synchronized %d skill(s) to %d global target(s)\n", len(names), len(catalog.Targets))
	return err
}

func replacementPrompt(input io.Reader, output io.Writer) skills.ConfirmReplacement {
	return func(name, target string) (bool, error) {
		if _, err := fmt.Fprintf(output, "Skill %q already exists in %s. Replace it? [y/N] ", name, target); err != nil {
			return false, err
		}
		var answer string
		if _, err := fmt.Fscanln(input, &answer); err != nil {
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		return strings.EqualFold(strings.TrimSpace(answer), "y") || strings.EqualFold(strings.TrimSpace(answer), "yes"), nil
	}
}

func removeSkill(name string, stdout io.Writer) error {
	catalog, err := skills.DefaultCatalog()
	if err != nil {
		return err
	}
	if err := catalog.Remove(name); err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "Removed skill %s from the catalog and managed targets\n", name)
	return err
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return "none"
	}
	result := names[0]
	for _, name := range names[1:] {
		result += ", " + name
	}
	return result
}
