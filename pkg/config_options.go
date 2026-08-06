package pkg

import (
	"fmt"
	"sort"
	"strings"
)

type ConfigOption struct {
	Topic       string
	Path        string
	Example     string
	Description string
}

var configOptions = []ConfigOption{
	{Topic: "core", Path: "[openscadgen].name", Example: `name = "sample_part"`, Description: "Human-readable project name used in reports and generated output."},
	{Topic: "core", Path: "[openscadgen].description", Example: `description = "Broom handle clip"`, Description: "Optional short project description shown in UI and generated docs."},
	{Topic: "core", Path: "[openscadgen].version", Example: `version = "v0.1"`, Description: "Version label used when building export paths."},
	{Topic: "core", Path: "[openscadgen].export_name_format", Example: `export_name_format = "{designFileName}_{paramSet}"`, Description: "Template for naming generated export folders and files."},
	{Topic: "paths", Path: "[openscadgen].input_path", Example: `input_path = "./part.scad"`, Description: "Primary OpenSCAD file to process, relative to the config file."},
	{Topic: "paths", Path: "[openscadgen].input_paths", Example: `[[openscadgen.input_paths]]\npath = "./part.scad"`, Description: "List of input files or variations to process as a batch."},
	{Topic: "paths", Path: "[openscadgen].output_path", Example: `output_path = "./export"`, Description: "Base output folder for generated artifacts."},
	{Topic: "generation", Path: "[openscadgen].run_type", Example: `run_type = "appendOrOverwrite"`, Description: "Controls whether exports are recreated or appended during processing."},
	{Topic: "generation", Path: "[openscadgen].no_part_id_letter", Example: `no_part_id_letter = true`, Description: "Suppresses automatic part ID lettering in generated names."},
	{Topic: "generation", Path: "[openscadgen].dont_use_manifold", Example: `dont_use_manifold = true`, Description: "Skips manifold-based export behavior when enabled."},
	{Topic: "open-scad", Path: "[openscadgen].custom_openscad_output_format", Example: `custom_openscad_output_format = "stl"`, Description: "Overrides the OpenSCAD output format used for exports."},
	{Topic: "open-scad", Path: "[openscadgen].custom_openscad_args", Example: `custom_openscad_args = "--disable=surfaces"`, Description: "Additional raw arguments passed to OpenSCAD."},
	{Topic: "open-scad", Path: "[openscadgen].export_image_quality", Example: `export_image_quality = "hq"`, Description: "Controls the image quality preset for generated preview images."},
	{Topic: "params", Path: "[openscadgen].global_params", Example: `global_params = { wall = 2.4 }`, Description: "Parameters applied to every generated instance unless overridden."},
	{Topic: "params", Path: "[openscadgen].param_sets", Example: `[[openscadgen.param_sets]]\nname = "wide"\nparams = { width = 120 }`, Description: "Named parameter bundles that can be reused across instances."},
	{Topic: "params", Path: "[openscadgen].instances", Example: `[[openscadgen.instances]]\nname = "default"\nparams = { width = 100 }`, Description: "Instance definitions to generate one output per parameter combination."},
	{Topic: "images", Path: "[openscadgen].images", Example: `[[openscadgen.images]]\nname = "front"\ncoord = "0,0,0"`, Description: "Camera definitions for preview image exports."},
	{Topic: "images", Path: "[openscadgen].export_text_sizing", Example: `[[openscadgen.export_text_sizing]]\nfont = "Roboto"\nparam_name = "label"`, Description: "Text sizing overrides used when rendering text-heavy exports."},
	{Topic: "images", Path: "[openscadgen.images].name", Example: `name = "front"`, Description: "Name of a preview camera definition."},
	{Topic: "images", Path: "[openscadgen.images].coord", Example: `coord = "0,0,0"`, Description: "Camera coordinates for a preview image."},
	{Topic: "images", Path: "[openscadgen.images].image_size", Example: `image_size = "1200x1200"`, Description: "Optional per-camera image size override."},
	{Topic: "images", Path: "[openscadgen.images].param_filter", Example: `param_filter = { finish = "matte" }`, Description: "Only use this camera when the matching parameters are present."},
	{Topic: "images", Path: "[openscadgen.export_text_sizing].font", Example: `font = "Roboto"`, Description: "Font family used for text sizing exports."},
	{Topic: "images", Path: "[openscadgen.export_text_sizing].param_name", Example: `param_name = "label"`, Description: "Parameter name that the text sizing entry applies to."},
	{Topic: "params", Path: "[openscadgen.param_sets].name", Example: `name = "wide"`, Description: "Name of the reusable parameter set."},
	{Topic: "params", Path: "[openscadgen.param_sets].params", Example: `params = { width = 120 }`, Description: "Parameter values contained in a reusable parameter set."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].path", Example: `path = "./part.scad"`, Description: "Path to an OpenSCAD file within an input_paths entry."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].raw_openscad_file", Example: `raw_openscad_file = "module body() { ... }"`, Description: "Inline OpenSCAD source content for a configured input path."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].raw_openscad_file_name", Example: `raw_openscad_file_name = "generated.scad"`, Description: "Filename to use when the OpenSCAD source is provided inline."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].export_name_format", Example: `export_name_format = "{designFileName}_{paramSet}"`, Description: "Overrides the export naming template for a specific input path."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].param_sets", Example: `param_sets = "wide,tall"`, Description: "Comma-separated param set names to apply to an input path."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].params", Example: `params = { width = 100 }`, Description: "Per-input parameter values merged into the instance."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].skip_images", Example: `skip_images = true`, Description: "Disables image generation for the specific input path."},
	{Topic: "inputs", Path: "[openscadgen.input_paths].ignore_param_when_processing", Example: `ignore_param_when_processing = "note"`, Description: "Excludes specific parameters from processing behavior."},
	{Topic: "instances", Path: "[openscadgen.instances].name", Example: `name = "left"`, Description: "Instance name used in generated file names and labels."},
	{Topic: "instances", Path: "[openscadgen.instances].description", Example: `description = "Left-handed variant"`, Description: "Optional description for a specific instance."},
	{Topic: "instances", Path: "[openscadgen.instances].params", Example: `params = { side = "left" }`, Description: "Per-instance parameter values."},
	{Topic: "instances", Path: "[openscadgen.instances].params_numbered", Example: `params_numbered = { slots = [1, 2, 3] }`, Description: "Parameters that should be expanded with numbering in output names."},
	{Topic: "instances", Path: "[openscadgen.instances].ignore_comma_in_params", Example: `ignore_comma_in_params = ["label"]`, Description: "Parameters treated as literal values even when they contain commas."},
	{Topic: "instances", Path: "[openscadgen.instances].direct_array_params", Example: `direct_array_params = ["points"]`, Description: "Parameters that should be passed through as direct arrays."},
	{Topic: "instances", Path: "[openscadgen.instances].images", Example: `[[openscadgen.instances.images]]\nname = "front"`, Description: "Instance-specific image exports."},
	{Topic: "instances", Path: "[openscadgen.instances].skip_images", Example: `skip_images = true`, Description: "Skips image generation for this instance."},
	{Topic: "instances", Path: "[openscadgen.instances].param_sets", Example: `param_sets = "wide,tall"`, Description: "Comma-separated param set names to apply to this instance."},
	{Topic: "instances", Path: "[openscadgen.instances.images].name", Example: `name = "front"`, Description: "Name of an instance-specific export camera."},
	{Topic: "instances", Path: "[openscadgen.instances.images].coord", Example: `coord = "0,0,0"`, Description: "Camera coordinates for an instance-specific export image."},
	{Topic: "instances", Path: "[openscadgen.instances.images].image_size", Example: `image_size = "1200x1200"`, Description: "Optional image size override for a per-instance camera."},
	{Topic: "instances", Path: "[openscadgen.instances.images].param_filter", Example: `param_filter = { side = "left" }`, Description: "Filters a per-instance camera to matching parameters only."},
}

func ListConfigOptions(topic string) []ConfigOption {
	topic = strings.TrimSpace(strings.ToLower(topic))
	out := make([]ConfigOption, 0, len(configOptions))
	for _, opt := range configOptions {
		if topic != "" && !strings.Contains(strings.ToLower(opt.Topic), topic) && !strings.Contains(strings.ToLower(opt.Path), topic) {
			continue
		}
		out = append(out, opt)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Topic == out[j].Topic {
			return out[i].Path < out[j].Path
		}
		return out[i].Topic < out[j].Topic
	})
	return out
}

func RenderConfigOptionsCLI(topic string) string {
	var b strings.Builder
	opts := ListConfigOptions(topic)
	if len(opts) == 0 {
		return fmt.Sprintf("%sConfig options%s\nNo options matched topic %q.\n", colorCyan, colorReset, topic)
	}
	currentTopic := ""
	for _, opt := range opts {
		if opt.Topic != currentTopic {
			currentTopic = opt.Topic
			fmt.Fprintf(&b, "\n%s%s%s\n", colorPurple, strings.ToUpper(currentTopic), colorReset)
		}
		fmt.Fprintf(&b, "%s%s%s\n", colorGreen, opt.Path, colorReset)
		fmt.Fprintf(&b, "  %sExample:%s\n    %s\n", colorYellow, colorReset, opt.Example)
		fmt.Fprintf(&b, "  %sDescription:%s %s\n\n", colorBlue, colorReset, opt.Description)
	}
	return strings.TrimSpace(b.String()) + "\n"
}

func ConfigOptionPaths() []string {
	opts := ListConfigOptions("")
	paths := make([]string, 0, len(opts))
	for _, opt := range opts {
		paths = append(paths, opt.Path)
	}
	return paths
}
