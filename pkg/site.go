package pkg

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kiwikid/openscadgen/pkg/models"
)

func BuildSite(repoRoot string, outputDir string, cmdFlags models.CmdFlags) error {
	ignoreSections := parseBuildSiteIgnoreSections(cmdFlags.BuildSiteIgnoreSections)
	for _, root := range []string{"examples", "bols2otropolis"} {
		scanRoot := filepath.Join(repoRoot, root)
		configs, err := ScanFolderForConfigFiles(scanRoot)
		if err != nil {
			return fmt.Errorf("scan %s: %w", root, err)
		}
		for _, cfg := range configs {
			if shouldSkipBuildSiteConfig(cfg.Path, ignoreSections) {
				continue
			}
			cmdFlags.ConfigFile = buildSiteConfigFilePath(scanRoot, cfg.Path)
			config, _, err := LoadConfigFromFile(cmdFlags)
			if err != nil {
				return fmt.Errorf("load config %s: %w", cfg.Path, err)
			}
			progress := &NoopProgress{}
			if _, err := Process(config, progress, nil, Operations{
				GenerateReport: true,
				ReportMode:     "pages",
			}, false); err != nil {
				return fmt.Errorf("process config %s: %w", cfg.Path, err)
			}
		}
	}
	if err := BuildPagesSite(repoRoot, outputDir); err != nil {
		return fmt.Errorf("assemble pages site: %w", err)
	}
	return nil
}

func buildSiteConfigFilePath(scanRoot string, configPath string) string {
	return filepath.Join(scanRoot, configPath)
}

func parseBuildSiteIgnoreSections(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	sections := make([]string, 0, len(parts))
	for _, part := range parts {
		section := strings.TrimSpace(part)
		if section != "" {
			sections = append(sections, section)
		}
	}
	return sections
}

func shouldSkipBuildSiteConfig(configPath string, ignoreSections []string) bool {
	if len(ignoreSections) == 0 {
		return false
	}
	normalized := filepath.ToSlash(configPath)
	for _, section := range ignoreSections {
		if section == "" {
			continue
		}
		if strings.Contains(normalized, section) {
			return true
		}
	}
	return false
}
