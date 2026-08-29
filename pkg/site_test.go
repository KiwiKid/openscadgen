package pkg

import (
	"path/filepath"
	"testing"
)

func TestBuildSiteConfigFilePath(t *testing.T) {
	scanRoot := filepath.Join("/tmp", "repo", "examples")
	got := buildSiteConfigFilePath(scanRoot, filepath.Join("rounded_button_box", "config.toml"))
	want := filepath.Join("/tmp", "repo", "examples", "rounded_button_box", "config.toml")
	if got != want {
		t.Fatalf("buildSiteConfigFilePath() = %q, want %q", got, want)
	}
}

func TestShouldSkipBuildSiteConfig(t *testing.T) {
	tests := []struct {
		name           string
		configPath     string
		ignoreSections []string
		want           bool
	}{
		{
			name:           "matches subtree",
			configPath:     filepath.Join("examples", "rounded_button_box", "config.toml"),
			ignoreSections: []string{"rounded_button_box"},
			want:           true,
		},
		{
			name:           "matches nested subtree",
			configPath:     filepath.Join("examples", "nested", "stablizer_foot", "config.toml"),
			ignoreSections: []string{"stablizer_foot"},
			want:           true,
		},
		{
			name:           "does not match",
			configPath:     filepath.Join("examples", "hinge_hanging", "config.toml"),
			ignoreSections: []string{"rounded_button_box", "stablizer_foot"},
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSkipBuildSiteConfig(tt.configPath, tt.ignoreSections)
			if got != tt.want {
				t.Fatalf("shouldSkipBuildSiteConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
