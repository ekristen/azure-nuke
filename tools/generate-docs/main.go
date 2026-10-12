package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/iancoleman/strcase"

	"github.com/ekristen/libnuke/pkg/docs"
	"github.com/ekristen/libnuke/pkg/registry"

	_ "github.com/ekristen/azure-nuke/resources"
)

func main() {
	registrations := registry.GetRegistrations()

	for _, reg := range registrations {
		if reg.Resource == nil {
			continue
		}

		propMap := docs.GeneratePropertiesMap(reg.Resource)

		niceName := strings.Join(camelcase.Split(reg.Name), " ")

		markdown := `# ` + niceName + "\n\n"
		markdown += "## Details\n\n"

		markdown += fmt.Sprintf("- **Type:** `%s`\n", reg.Name)
		markdown += fmt.Sprintf("- **Scope:** %s\n", reg.Scope)

		markdown += "\n"

		markdown += "## Properties\n\n"
		var propKeys []string
		for k := range propMap {
			propKeys = append(propKeys, k)
		}
		sort.Strings(propKeys)
		for _, k := range propKeys {
			v := propMap[k]
			if v == "" {
				v = "No description provided"
			}
			markdown += fmt.Sprintf("- **`%s`**: %s\n", k, v)
		}

		if len(reg.DependsOn) > 0 {
			markdown += "## Depends On\n\n"
			markdown += "!!! Experimental Feature\n"
			markdown += "    This is an **experimental** feature, please read more about it here <>. This feature " +
				"attempts to remove all resources in one resource type before moving onto the " +
				"dependent resource type\n\n"
			for _, dep := range reg.DependsOn {
				depName := strings.Join(camelcase.Split(dep), " ")
				depFilename := strcase.ToKebab(dep)
				markdown += fmt.Sprintf("- [%s](%s.md)\n", depName, depFilename)
			}
		}

		if len(reg.Settings) > 0 {
			markdown += "## Settings\n\n"
			for _, v := range reg.Settings {
				markdown += fmt.Sprintf("- `%s`\n", v)
			}
		}

		if len(reg.DeprecatedAliases) > 0 {
			markdown += "## Deprecated Aliases\n\n"
			markdown += "These are deprecated aliases for the resource type, usually misspellings or old names " +
				"that have been replaced with a new resource type.\n\n"
			for _, alias := range reg.DeprecatedAliases {
				aliasName := strings.Join(camelcase.Split(alias), " ")
				aliasFilename := strcase.ToKebab(alias)
				markdown += fmt.Sprintf("- [%s](%s.md)\n", aliasName, aliasFilename)
			}
		}

		filename := strcase.ToKebab(reg.Name)

		_ = os.WriteFile(fmt.Sprintf("docs/resources/%s.md", filename), []byte(markdown), 0o644) //nolint: gosec
	}

	var regs []string
	for _, reg := range registrations {
		regs = append(regs, reg.Name)
	}

	sort.Strings(regs)

	for _, reg := range regs {
		name := strings.Join(camelcase.Split(reg), " ")
		filename := strcase.ToKebab(reg)
		fmt.Printf("- %s: resources/%s.md\n", name, filename)
	}
}
