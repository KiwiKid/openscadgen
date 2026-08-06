package pkg

import "testing"

func TestConfigOptionPathsCoverSchema(t *testing.T) {
	want := map[string]struct{}{
		"[openscadgen].name": {},
		"[openscadgen].description": {},
		"[openscadgen].input_path": {},
		"[openscadgen].input_paths": {},
		"[openscadgen].output_path": {},
		"[openscadgen].version": {},
		"[openscadgen].no_part_id_letter": {},
		"[openscadgen].run_type": {},
		"[openscadgen].export_name_format": {},
		"[openscadgen].global_params": {},
		"[openscadgen].param_sets": {},
		"[openscadgen].custom_openscad_output_format": {},
		"[openscadgen].custom_openscad_args": {},
		"[openscadgen].export_image_quality": {},
		"[openscadgen].images": {},
		"[openscadgen].export_text_sizing": {},
		"[openscadgen].instances": {},
		"[openscadgen].dont_use_manifold": {},
		"[openscadgen.input_paths].path": {},
		"[openscadgen.input_paths].raw_openscad_file": {},
		"[openscadgen.input_paths].raw_openscad_file_name": {},
		"[openscadgen.input_paths].export_name_format": {},
		"[openscadgen.input_paths].param_sets": {},
		"[openscadgen.input_paths].params": {},
		"[openscadgen.input_paths].skip_images": {},
		"[openscadgen.input_paths].ignore_param_when_processing": {},
		"[openscadgen.param_sets].name": {},
		"[openscadgen.param_sets].params": {},
		"[openscadgen.images].name": {},
		"[openscadgen.images].coord": {},
		"[openscadgen.images].image_size": {},
		"[openscadgen.images].param_filter": {},
		"[openscadgen.export_text_sizing].font": {},
		"[openscadgen.export_text_sizing].param_name": {},
		"[openscadgen.instances].name": {},
		"[openscadgen.instances].description": {},
		"[openscadgen.instances].params": {},
		"[openscadgen.instances].param_sets": {},
		"[openscadgen.instances].params_numbered": {},
		"[openscadgen.instances].ignore_comma_in_params": {},
		"[openscadgen.instances].direct_array_params": {},
		"[openscadgen.instances].images": {},
		"[openscadgen.instances].skip_images": {},
		"[openscadgen.instances.images].name": {},
		"[openscadgen.instances.images].coord": {},
		"[openscadgen.instances.images].image_size": {},
		"[openscadgen.instances.images].param_filter": {},
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
