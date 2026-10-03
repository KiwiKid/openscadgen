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

func presetImageDescription() string {
	return strings.Join([]string{
		"Camera definitions for preview image exports.",
		"Preset directions and coordinates:",
		"- top / down: 0,0,0,0,0,0,300",
		"- top-far: 0,0,0,0,0,0,800",
		"- bottom / up: 0,0,0,180,0,0,300",
		"- bottom-far: 0,0,0,180,0,0,800",
		"- front: 0,0,0,90,0,0,300",
		"- front-far: 0,0,0,90,0,0,800",
		"- back: 0,0,0,270,0,0,300",
		"- back-far: 0,0,0,270,0,0,800",
		"- side / left: 0,0,0,0,90,0,300",
		"- side-near / left-near: 0,0,0,0,90,0,150",
		"- left-far: 0,0,0,0,90,0,800",
		"- right: 0,0,0,0,0,0,300",
		"- right-far: 0,0,0,0,0,0,800",
		"- nice: 0,0,0,45,0,45,350",
		"- nice-near: 0,0,0,45,0,45,150",
		"- nice-far: 0,0,0,45,0,45,800",
		"- nice-100 through nice-1000: 0,0,0,45,0,45,<distance>",
	}, "\n")
}

var configOptions = []ConfigOption{
	{Topic: "core", Path: "[openscadgen].export_name_format", Example: `export_name_format = "{designFileName}_{paramSet}"`, Description: "Template for naming generated export folders and files."},
	{Topic: "core", Path: "[openscadgen].global_params", Example: `global_params = { wall = 2.4 }`, Description: "Parameters applied to every generated instance unless overridden."},
	{Topic: "core", Path: "[openscadgen].instances", Example: `[[openscadgen.instances]]\nname = "default"\nparams = { width = 100 }`, Description: "Instance definitions to generate one output per parameter combination."},
	{Topic: "core", Path: "[openscadgen].name", Example: `name = "sample_part"`, Description: "Human-readable project name used in reports and generated output."},
	{Topic: "core", Path: "[openscadgen].version", Example: `version = "v0.1"`, Description: "Version label used when building export paths."},
	{Topic: "core", Path: "[openscadgen].description", Example: `description = "Broom handle clip"`, Description: "Optional short project description shown in UI and generated docs."},
	{Topic: "core", Path: "[openscadgen].input_path", Example: `input_path = "./part.scad"`, Description: "Primary OpenSCAD file to process, relative to the config file."},
	{Topic: "core", Path: "[openscadgen].favicon", Example: `favicon = "./favicon.png"`, Description: "Optional favicon image, relative to config.toml. It is displayed inside openscadgen's circular favicon in web and offline reports."},
	{Topic: "core", Path: "[openscadgen].input_paths", Example: `[[openscadgen.input_paths]]\npath = "./part.scad"`, Description: "List of input files or variations to process as a batch."},
	{Topic: "core", Path: "[openscadgen].output_path", Example: `output_path = "./export"`, Description: "Base output folder for generated artifacts."},
	{Topic: "advanced", Path: "[openscadgen].param_sets", Example: `[[openscadgen.param_sets]]\nname = "wide"\nparams = { width = 120 }`, Description: "Named parameter bundles that can be reused across instances."},
	{Topic: "advanced", Path: "[openscadgen].run_type", Example: `run_type = "appendOrOverwrite"`, Description: "Controls whether exports are recreated or appended during processing."},
	{Topic: "advanced", Path: "[openscadgen].no_part_id_letter", Example: `no_part_id_letter = true`, Description: "Suppresses automatic part ID lettering in generated names."},
	{Topic: "advanced", Path: "[openscadgen].dont_use_manifold", Example: `dont_use_manifold = true`, Description: "Skips manifold-based export behavior when enabled."},
	{Topic: "advanced", Path: "[openscadgen].custom_openscad_output_format", Example: `custom_openscad_output_format = "stl"`, Description: "Overrides the OpenSCAD output format used for exports."},
	{Topic: "advanced", Path: "[openscadgen].custom_openscad_args", Example: `custom_openscad_args = "--disable=surfaces"`, Description: "Additional raw arguments passed to OpenSCAD."},
	{Topic: "advanced", Path: "[openscadgen].export_image_quality", Example: `export_image_quality = "hq"`, Description: "Controls the image quality preset for generated preview images."},
	{Topic: "advanced", Path: "[openscadgen].images", Example: `[[openscadgen.images]]\nname = "front"\ncoord = "0,0,0"`, Description: presetImageDescription()},
	{Topic: "advanced", Path: "[openscadgen].image", Example: `[[openscadgen.image]]\nname = "nice"\nparams = { mode = "assembled" }`, Description: "One targeted image render with an explicit parameter set. It reuses exactly one matching instance, or creates an image-only instance for the renderer and report."},
	{Topic: "advanced", Path: "[openscadgen].tests", Example: `[[openscadgen.tests]]\ntype = "is-flat"\nside = "bottom"\nlevel = "medium"`, Description: "Checks applied to generated parts or renders. is-flat verifies a meaningful flat contact area; no-holes rejects enclosed image background; is-different renders two parameter variants and requires visible differences; min-color-pixels requires a visible coloured area; color-in-bounding-box requires a colour inside a square image region and writes a red/green diagnostic overlay."},
	{Topic: "advanced", Path: "[openscadgen.tests].camera", Example: `camera = "top"`, Description: "Required for no-holes. The preset or named camera to render and inspect."},
	{Topic: "advanced", Path: "[openscadgen.tests].image_size", Example: `image_size = "1200x1200"`, Description: "Optional render size for the no-holes test camera."},
	{Topic: "advanced", Path: "[openscadgen.tests].min_hole_pixels", Example: `min_hole_pixels = 4`, Description: "Smallest enclosed background component that fails a no-holes test. Defaults to 1."},
	{Topic: "advanced", Path: "[openscadgen.tests].instance_target", Example: `instance_target = "preview"`, Description: "Required for is-different. Named [[openscadgen.instances]] entry whose parameters are inherited by both rendered variants."},
	{Topic: "advanced", Path: "[openscadgen.tests].params", Example: `params = { enabled = false }`, Description: "Parameter overrides for the first is-different image; unspecified values inherit from instance_target."},
	{Topic: "advanced", Path: "[openscadgen.tests].different_params", Example: `different_params = { enabled = true }`, Description: "Parameter overrides for the second is-different image; unspecified values inherit from instance_target."},
	{Topic: "advanced", Path: "[openscadgen.tests].min_different_pixels", Example: `min_different_pixels = 10`, Description: "Minimum visibly different pixels required by is-different. Defaults to 1."},
	{Topic: "advanced", Path: "[openscadgen.tests].color", Example: `color = "#ff0000"`, Description: "Required #RRGGBB target colour for a min-color-pixels image test."},
	{Topic: "advanced", Path: "[openscadgen.tests].min_color_pixels", Example: `min_color_pixels = 100`, Description: "Minimum pixels near the target colour required by min-color-pixels. Defaults to 1."},
	{Topic: "advanced", Path: "[openscadgen.tests].bounding_box", Example: `bounding_box = [120, 80, 64]`, Description: "Required for color-in-bounding-box: [left, top, size] in exported-image pixels. Matching tests for one image are combined in one green/red diagnostic overlay."},
	{Topic: "advanced", Path: "[openscadgen].export_text_sizing", Example: `[[openscadgen.export_text_sizing]]\nfont = "Roboto"\nparam_name = "label"`, Description: "Text sizing overrides used when rendering text-heavy exports."},
	{Topic: "advanced", Path: "[openscadgen.images].name", Example: `name = "front"`, Description: "Name of a preview camera definition."},
	{Topic: "advanced", Path: "[openscadgen.images].type", Example: `type = "all_instances"`, Description: "Optional aggregate image type. all_instances creates one PNG montage from this camera's successful renders across every matching instance."},
	{Topic: "advanced", Path: "[openscadgen.images].export_name_format", Example: `export_name_format = "previews/{designFileName}"`, Description: "Overrides this camera's path below img/<version>. Supports the same placeholders as the main export_name_format."},
	{Topic: "advanced", Path: "[openscadgen.images].coord", Example: `coord = "0,0,0"`, Description: "Camera coordinates for a preview image. Use preset directions like top, front, back, left, right, side, and nice, or provide a custom 7-value OpenSCAD camera string."},
	{Topic: "advanced", Path: "[openscadgen.images].image_size", Example: `image_size = "1200x1200"`, Description: "Optional per-camera image size override."},
	{Topic: "advanced", Path: "[openscadgen.images].param_filter", Example: `param_filter = { finish = "matte" }`, Description: "Only use this camera when the matching parameters are present."},
	{Topic: "advanced", Path: "[openscadgen.image].name", Example: `name = "nice"`, Description: "Camera preset name for the single preview image."},
	{Topic: "advanced", Path: "[openscadgen.image].instance_name", Example: `instance_name = "hero-preview"`, Description: "Optional report name when this image needs an image-only instance."},
	{Topic: "advanced", Path: "[openscadgen.image].input_path", Example: `input_path = "./part.scad"`, Description: "Required when the design has more than one input path."},
	{Topic: "advanced", Path: "[openscadgen.image].params", Example: `params = { mode = "assembled" }`, Description: "Parameters for this one image render. They must identify at most one configured instance."},
	{Topic: "advanced", Path: "[openscadgen.image].param_sets", Example: `param_sets = "assembled"`, Description: "Comma-separated parameter sets to apply to this image render."},
	{Topic: "advanced", Path: "[openscadgen.image].export_name_format", Example: `export_name_format = "{designFileName}_hero"`, Description: "Optional naming format for an image-only report instance."},
	{Topic: "advanced", Path: "[openscadgen.image].coord", Example: `coord = "0,0,0,45,0,45,350"`, Description: "Optional custom 7-value OpenSCAD camera string; omit it to use the name preset."},
	{Topic: "advanced", Path: "[openscadgen.image].image_size", Example: `image_size = "1200x1200"`, Description: "Optional dimensions for this image."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].path", Example: `path = "./part.scad"`, Description: "Path to an OpenSCAD file within an input_paths entry."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].raw_openscad_file", Example: `raw_openscad_file = "module body() { ... }"`, Description: "Inline OpenSCAD source content for a configured input path."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].raw_openscad_file_name", Example: `raw_openscad_file_name = "generated.scad"`, Description: "Filename to use when the OpenSCAD source is provided inline."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].export_name_format", Example: `export_name_format = "{designFileName}_{paramSet}"`, Description: "Overrides the export naming template for a specific input path."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].param_sets", Example: `param_sets = "wide,tall"`, Description: "Comma-separated param set names to apply to an input path."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].params", Example: `params = { width = 100 }`, Description: "Per-input parameter values merged into the instance."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].skip_images", Example: `skip_images = true`, Description: "Disables image generation for the specific input path."},
	{Topic: "advanced", Path: "[openscadgen.input_paths].ignore_param_when_processing", Example: `ignore_param_when_processing = "note"`, Description: "Excludes specific parameters from processing behavior."},
	{Topic: "advanced", Path: "[openscadgen.param_sets].name", Example: `name = "wide"`, Description: "Name of the reusable parameter set."},
	{Topic: "advanced", Path: "[openscadgen.param_sets].params", Example: `params = { width = 120 }`, Description: "Parameter values contained in a reusable parameter set."},
	{Topic: "advanced", Path: "[openscadgen.export_text_sizing].font", Example: `font = "Roboto"`, Description: "Font family used for text sizing exports."},
	{Topic: "advanced", Path: "[openscadgen.export_text_sizing].param_name", Example: `param_name = "label"`, Description: "Parameter name that the text sizing entry applies to."},
	{Topic: "advanced", Path: "[openscadgen.instances].name", Example: `name = "left"`, Description: "Instance name used in generated file names and labels."},
	{Topic: "advanced", Path: "[openscadgen.instances].description", Example: `description = "Left-handed variant"`, Description: "Optional description for a specific instance."},
	{Topic: "advanced", Path: "[openscadgen.instances].export_name_format", Example: `export_name_format = "{designFileName}_{name}"`, Description: "Overrides the export naming template for a specific instance."},
	{Topic: "advanced", Path: "[openscadgen.instances].params", Example: `params = { side = "left" }`, Description: "Per-instance parameter values."},
	{Topic: "advanced", Path: "[openscadgen.instances].param_sets", Example: `param_sets = "wide,tall"`, Description: "Comma-separated param set names to apply to this instance."},
	{Topic: "advanced", Path: "[openscadgen.instances].params_numbered", Example: `params_numbered = { slots = [1, 2, 3] }`, Description: "Parameters that should be expanded with numbering in output names."},
	{Topic: "advanced", Path: "[openscadgen.instances].ignore_comma_in_params", Example: `ignore_comma_in_params = ["label"]`, Description: "Parameters treated as literal values even when they contain commas."},
	{Topic: "advanced", Path: "[openscadgen.instances].direct_array_params", Example: `direct_array_params = ["points"]`, Description: "Parameters that should be passed through as direct arrays."},
	{Topic: "advanced", Path: "[openscadgen.instances].images", Example: `[[openscadgen.instances.images]]\nname = "front"`, Description: "Instance-specific image exports.\n\n" + presetImageDescription()},
	{Topic: "advanced", Path: "[openscadgen.instances].skip_images", Example: `skip_images = true`, Description: "Skips image generation for this instance."},
	{Topic: "advanced", Path: "[openscadgen.instances.images].name", Example: `name = "front"`, Description: "Name of an instance-specific export camera."},
	{Topic: "advanced", Path: "[openscadgen.instances.images].export_name_format", Example: `export_name_format = "previews/{designFileName}"`, Description: "Overrides this camera's path below img/<version>. Supports the same placeholders as the main export_name_format."},
	{Topic: "advanced", Path: "[openscadgen.instances.images].coord", Example: `coord = "0,0,0"`, Description: "Camera coordinates for an instance-specific export image. Use preset directions like top, front, back, left, right, side, and nice, or provide a custom 7-value OpenSCAD camera string."},
	{Topic: "advanced", Path: "[openscadgen.instances.images].image_size", Example: `image_size = "1200x1200"`, Description: "Optional image size override for a per-instance camera."},
	{Topic: "advanced", Path: "[openscadgen.instances.images].param_filter", Example: `param_filter = { side = "left" }`, Description: "Filters a per-instance camera to matching parameters only."},
	{Topic: "advanced", Path: "[openscadgen.instances.images].tests", Example: `[[openscadgen.instances.images.tests]]\ntype = "color-in-bounding-box"\ncolor = "#ff0000"\nmin_color_pixels = 1\nbounding_box = [120, 80, 64]`, Description: "Tests evaluated only against this instance image. Use color-in-bounding-box to require or forbid a colour within a selected square region."},
}

func topicRank(topic string) int {
	switch strings.ToLower(strings.TrimSpace(topic)) {
	case "core":
		return 0
	case "advanced":
		return 1
	default:
		return 2
	}
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
	sort.SliceStable(out, func(i, j int) bool {
		if topicRank(out[i].Topic) == topicRank(out[j].Topic) {
			if out[i].Topic == out[j].Topic {
				return out[i].Path < out[j].Path
			}
			return out[i].Topic < out[j].Topic
		}
		return topicRank(out[i].Topic) < topicRank(out[j].Topic)
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
