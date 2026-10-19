package server

import (
	"github.com/BurntSushi/toml"
	"github.com/go-playground/validator/v10"
	"github.com/kiwikid/openscadgen/pkg/models"
	"testing"
)

func TestConfigLinks(t *testing.T) {
	const source = `[openscadgen]
name = "Example"
[attribution]
url = "https://example.com/original"
label = "Original design"
author = "Designer"
subtitle = "Adapted with permission"
[[publication]]
url = "https://www.printables.com/model/123"
label = "Printables"
[[publication]]
url = "https://example.com/makerlab/123"
`
	var config models.Config
	metadata, err := toml.Decode(source, &config)
	if err != nil {
		t.Fatal(err)
	}
	if len(metadata.Undecoded()) != 0 {
		t.Fatalf("unrecognized fields: %v", metadata.Undecoded())
	}
	if err := validator.New().Struct(config); err != nil {
		t.Fatal(err)
	}
	if config.Attribution.Author != "Designer" || len(config.Publication) != 2 {
		t.Fatalf("unexpected links: %+v", config)
	}
	sections, _ := buildEditPreviewSections(&config, nil)
	for _, kind := range []string{"attribution", "publication"} {
		found := false
		for _, section := range sections {
			if section.Title == kind {
				found = true
				if !section.HasItems || section.Items[0].Title == "" {
					t.Fatalf("missing link preview: %+v", section)
				}
			}
		}
		if !found {
			t.Fatalf("missing %s section", kind)
		}
	}
	config.Publication[1].URL = ""
	if err := validator.New().Struct(config); err == nil {
		t.Fatal("missing publication URL accepted")
	}
	config.Publication[1].URL = "javascript:alert(1)"
	if err := validator.New().Struct(config); err == nil {
		t.Fatal("non-HTTP URL accepted")
	}
}
