package pkg

import (
	"strings"
	"testing"
)

func TestConfigOptionPathsCoverSchema(t *testing.T) {
	want := map[string]struct{}{
		"[openscadgen].name":                                     {},
		"[openscadgen].description":                              {},
		"[openscadgen].input_path":                               {},
		"[openscadgen].input_paths":                              {},
		"[openscadgen].output_path":                              {},
		"[openscadgen].version":                                  {},
		"[openscadgen].no_part_id_letter":                        {},
		"[openscadgen].run_type":                                 {},
		"[openscadgen].export_name_format":                       {},
		"[openscadgen].global_params":                            {},
		"[openscadgen].param_sets":                               {},
		"[openscadgen].custom_openscad_output_format":            {},
		"[openscadgen].custom_openscad_args":                     {},
		"[openscadgen].export_image_quality":                     {},
		"[openscadgen].images":                                   {},
		"[openscadgen].export_text_sizing":                       {},
		"[openscadgen].instances":                                {},
		"[openscadgen].dont_use_manifold":                        {},
		"[openscadgen.input_paths].path":                         {},
		"[openscadgen.input_paths].raw_openscad_file":            {},
		"[openscadgen.input_paths].raw_openscad_file_name":       {},
		"[openscadgen.input_paths].export_name_format":           {},
		"[openscadgen.input_paths].param_sets":                   {},
		"[openscadgen.input_paths].params":                       {},
		"[openscadgen.input_paths].skip_images":                  {},
		"[openscadgen.input_paths].ignore_param_when_processing": {},
		"[openscadgen.param_sets].name":                          {},
		"[openscadgen.param_sets].params":                        {},
		"[openscadgen.images].name":                              {},
		"[openscadgen.images].type":                              {},
		"[openscadgen.images].export_name_format":                {},
		"[openscadgen.images].coord":                             {},
		"[openscadgen.images].image_size":                        {},
		"[openscadgen.images].param_filter":                      {},
		"[openscadgen.export_text_sizing].font":                  {},
		"[openscadgen.export_text_sizing].param_name":            {},
		"[openscadgen.instances].name":                           {},
		"[openscadgen.instances].description":                    {},
		"[openscadgen.instances].export_name_format":             {},
		"[openscadgen.instances].params":                         {},
		"[openscadgen.instances].param_sets":                     {},
		"[openscadgen.instances].params_numbered":                {},
		"[openscadgen.instances].ignore_comma_in_params":         {},
		"[openscadgen.instances].direct_array_params":            {},
		"[openscadgen.instances].images":                         {},
		"[openscadgen.instances].skip_images":                    {},
		"[openscadgen.instances.images].name":                    {},
		"[openscadgen.instances.images].export_name_format":      {},
		"[openscadgen.instances.images].coord":                   {},
		"[openscadgen.instances.images].image_size":              {},
		"[openscadgen.instances.images].param_filter":            {},
	}

	got := map[string]struct{}{}
	for _, path := range ConfigOptionPaths() {
		got[path] = struct{}{}
	}

	for path := range want {
		if _, ok := got[path]; !ok {
			t.Fatalf("missing config option path: %s", path)
		}
	}
	for path := range got {
		if _, ok := want[path]; !ok {
			t.Fatalf("unexpected config option path: %s", path)
		}
	}
}

func TestConfigOptionsImagesDescriptionsIncludePresets(t *testing.T) {
	opts := ListConfigOptions("images")
	var imagesDesc, instanceImagesDesc, coordDesc string
	for _, opt := range opts {
		switch opt.Path {
		case "[openscadgen].images":
			imagesDesc = opt.Description
		case "[openscadgen.instances].images":
			instanceImagesDesc = opt.Description
		case "[openscadgen.images].coord":
			coordDesc = opt.Description
		}
	}

	if !strings.Contains(imagesDesc, "top / down: 0,0,0,0,0,0,300") {
		t.Fatalf("expected top preset details in [openscadgen].images description, got %q", imagesDesc)
	}
	if !strings.Contains(instanceImagesDesc, "nice-100 through nice-1000") {
		t.Fatalf("expected nice preset range in [openscadgen.instances].images description, got %q", instanceImagesDesc)
	}
	if !strings.Contains(coordDesc, "custom 7-value OpenSCAD camera string") {
		t.Fatalf("expected coord guidance in [openscadgen.images].coord description, got %q", coordDesc)
	}
}
