openscadgen is a simple tool for generating a specific a set of .stl files (or other openscad export formats) from one (or many!) openscad files and a simple human-readable toml config file.

> [!WARNING]
> openscadgen is designed to run locally, in a local-development-style environment. It is not ready for, or currently designed to be safely run as, an internet-exposed or production server. Hardening work is in progress, but it is not complete. Do not expose the server to untrusted networks or users.

The config file allows for parameter combinations, multiple input files, image generation, and generated a report detailing all the generations and the parameters used in each.

The goal of the tool is to ease the development, management, production and distribution of large numbers designs or design options based on single openscad file.

A simpler alterative to programmatic wrappers like Makefiles, PythonSCAD, AnchorSCAD, LuaCAD, etc - designed provides a more accessible/structured approach.

https://github.com/user-attachments/assets/61d605fa-3b5a-43c0-98d2-71357f96fc32


openscadgen includes two main modes of operation:
- Direct Running:
(run the generation for this config.toml file)
```
./openscadgen -c ./examples/football_cards/config.toml
```

- Server Mode:
(start a server and serve the projects in the folder "./examples/" at http://localhost:7203)
```
./openscadgen -sf ./examples/
```

Server mode binds to `127.0.0.1` by default. `--server-address` is an explicit
opt-in for a non-loopback bind; do not expose it directly to a network because
the application can modify project files and launch local programs.

## Prerequisites

Before using openscadgen, you need:

1. **OpenSCAD** - The tool requires OpenSCAD to be installed and available in your PATH
2. **A config file** - A `.toml` configuration file that defines your project
3. **OpenSCAD files** - The `.scad` files you want to generate variations of

## Compatibility

`openscadgen` ships native x86-64 and ARM64 binaries for the desktop platforms below. It invokes the `openscad` command on your `PATH`; check the renderer it will use with `openscad --version`.

| Platform | openscadgen | OpenSCAD releases | Known limitations |
| --- | --- | --- | --- |
| macOS (Intel, Apple Silicon) | Supported | Current stable and nightly snapshots | The server's OpenSCAD and BOSL2 install/upgrade buttons require Homebrew and are macOS-only. macOS may require allowing the downloaded binary in Privacy & Security. |
| Linux (x86-64, ARM64) | Supported | Ubuntu CI's packaged stable release; current nightly is also exercised by the release workflow | Install OpenSCAD and its EGL/OpenGL runtime libraries for headless rendering. Use your distribution's package manager to install or upgrade OpenSCAD. |
| Windows (x86-64, ARM64) | Supported | Current stable releases | Install OpenSCAD and manage libraries manually; the server's install/upgrade buttons are not available. Windows Security may require unblocking the downloaded executable. |
| Docker host (any OS supported by Docker) | Supported through the Linux container | The image supplies Debian's OpenSCAD package and BOSL2 | Local desktop integration, including opening a SCAD file in the host GUI, is not available from the container. |

OpenSCAD is responsible for parsing and rendering each model, so a model's own libraries, fonts, and features can impose stricter requirements than this matrix. Nightly builds are useful for newer renderer features, but can change behaviour; pin a known OpenSCAD version when reproducible exports matter.

## Quick Start

1. **Install openscadgen** from the [releases page](https://github.com/kiwikid/openscadgen/releases)
2. **Initialize a new project:**
   ```bash
   ./openscadgen -i ./examples/my-first-project
   ```
3. **Run the tool**:
   ```bash
   ./openscadgen -c config.toml
   ```

## Attribution and publication links

The config form supports one source credit and multiple published locations.
Use top-level tables outside `[openscadgen]`:

```toml
[attribution]
url = "https://example.com/original-design"
label = "Original design"
author = "Designer name"
subtitle = "Adapted from the original"

[[publication]]
url = "https://www.printables.com/model/123"
label = "Printables"

[[publication]]
url = "https://example.com/makerlab/design"
label = "MakerLab"
```

Each entry requires an HTTP or HTTPS `url`; `label`, `author`, and `subtitle`
are optional. In the config form, add or edit a link, apply it to the dry run,
then save the validated configuration.

## Basic Configuration

The simplest configuration uses:
-  one input scad file
and creates multiple .stl files based on a combination of params against each [[openscadgen.instances]] :

```toml
[openscadgen]
name = "simple-clip"
description = "A basic parametric clip"
input_path = "./clip.scad" # input_path is relative to the config.toml file
export_name_format = "clip-{width}mm-{height}mm"
version = "v1.0"

# This will run the ./clip.scad file 9 times (use export_name_format for file naming) following params:
# - width = "10" & height = "5"
# - width = "10" & height = "10"
# - width = "10" & height = "15"
# - width = "20" & height = "5"
# - width = "20" & height = "10"
# - width = "20" & height = "15"
# - width = "30" & height = "5"
# - width = "30" & height = "10"
# - width = "30" & height = "15"
[[openscadgen.instances]]
params = { width = "10,20,30", height = "5,10,15" }
```
When run via:
```
./openscadgen -c ./config.toml
```
openscadgen will:
- Use a single input file (`clip.scad`)
- Generate all combinations of width and height
- Export STLs at the export_name_format with the format `clip-10mm-5mm.stl`, `clip-10mm-10mm.stl`, etc.

## Multiple Input Files

When you need to generate multiple designs using the same parameters - multiple [[openscadgen.input_paths]] can be used

```toml
[openscadgen]
name = "cup-holder"
description = "A cup holder with multiple components"
export_name_format = "{designFileName}-{diameter}mm"
version = "v1.0"

[[openscadgen.input_paths]]
path = "./cup-holder-main.scad"

[[openscadgen.input_paths]]
path = "./cup-holder-adapter.scad"

[[openscadgen.instances]]
params = { diameter = "70,80,90", style = "minimal,decorative" }
```

Both designs are rendered using the same diameter parameters, with the adapter having its own export path format  - the filename of the source .scad files are replaced as {designFileName} in the `export_name_format`

## Adding Image Previews

To generate preview images of your STL files:

```toml
[openscadgen]
name = "phone-stand"
description = "A parametric phone stand"
input_path = "./phone-stand.scad"
export_name_format = "stand-{angle}deg"
version = "v1.0"

[[openscadgen.instances]]
params = { angle = "45,60,75" }

[[openscadgen.images]]
name = "top"

[[openscadgen.images]]
name = "front"

[[openscadgen.images]]
name = "side"
coord = "0,0,0,90,0,0,600"
```

This adds preview images from multiple angles for each generated STL.

To additionally create one montage that shows every instance from the same camera, set the image type to `all_instances`:

```toml
[[openscadgen.images]]
name = "nice"
type = "all_instances"
```

This keeps the individual `nice` previews and writes `all-instances-nice.png` in the export version folder.

## Using Parameter Sets

For more complex designs, parameter sets help organize related parameters:

```toml
[openscadgen]
name = "complex-bracket"
description = "A bracket with multiple configurations"
input_path = "./bracket.scad"
export_name_format = "bracket-{size}-{style}"
version = "v1.0"

[[openscadgen.param_sets]]
name = "small"
params = { width = "20", height = "30", thickness = "3" }

[[openscadgen.param_sets]]
name = "large"
params = { width = "40", height = "60", thickness = "5" }

[[openscadgen.param_sets]]
name = "minimal"
params = { style = "minimal", edge_radius = "2" }

[[openscadgen.param_sets]]
name = "decorative"
params = { style = "decorative", edge_radius = "5", pattern = "diamond" }


[[openscadgen.instances]]
name = "small-minimal"
# All the parameters from the param_sets listed here will be applied during generation
param_sets = "small,minimal"

[[openscadgen.instances]]
name = "large-decorative"
param_sets = "large,decorative"
```

This configuration:
- Defines reusable parameter sets
- Combines parameter sets to create different configurations
- Makes it easier to manage complex parameter combinations

## Key Concepts

### Global Parameters
Use `global_params` to set parameters that apply to all instances:

```toml
[openscadgen]
name = "my-design"
global_params = { partType = "box", material = "PLA" }
```

### File Name Placeholders
In `export_name_format`, you can use these placeholders:
- `{paramName}` - Value of any parameter
- `{designFileName}` - Name of the source .scad file (without extension)
- `{instanceName}` - Name of the instance (if specified)

### Output Structure
Files are organized as:
```
export/
└── v1.0/
    ├── design-param1-param2.stl
    ├── design-param1-param2-top.png
    └── report.html
```

The `report.html` file provides a visual overview of all generated files with images and metadata, making it easy to browse your results.

### Browser rendering (optional)

Add `-browser-report` (`-br`) to embed a browser renderer in the report:

```bash
./openscadgen -c config.toml -br
```

The report then includes a **Browser render** panel. It re-renders selected
instances with the bundled OpenSCAD WebAssembly build, using the same applied
parameters as the host run, and downloads the resulting STL files. The 3D
preview uses WebGL; OpenSCAD's geometry evaluation itself stays on the CPU.
The control panel and URL-state format follow the
[OpenSCAD Playground](https://github.com/openscad/openscad-playground)
where practical.

Notes:

- The browser renderer needs the `browser/` folder that openscadgen writes
  beside `report.html`. Serve the export folder over HTTP; most browsers block
  workers and WebAssembly from `file://`.
- The browser build cannot export PNG/PDF (no OpenGL/Cairo in OpenSCAD-WASM),
  so image previews and image-based tests remain host-generated.
- `browser/openscad.js` and `browser/openscad.wasm` are downloaded from a
  pinned OpenSCAD Playground build. In a source checkout, run
  `node scripts/fetch-browser-assets.mjs` once. Release archives include them.
- **Copy share link** stores the compressed config, design sources and resolved
  instances in the URL fragment (`#og=…`), not the query string. Query
  parameters are reserved for small overrides: `?instance=name` selects an
  instance, and `?config=<value>&design=<value>` (optionally `?source=path`)
  override the embedded config or source. For those two, a `1.`-prefixed value
  is gzip+base64 state; anything else is treated as plain text. Prefer the
  fragment for full designs because URLs have finite length and query strings
  are logged by servers and proxies.

Each normal processing run also refreshes `export/<version>/source.zip`. It
contains the exact `config.toml` and the configured `.scad` input files for
that version. The archive is compressed, replaced rather than accumulated on
later runs of the same version, and stays under `export/`, which project search
already excludes. It deliberately does not bundle external OpenSCAD libraries
or imported assets.

### Part Type Generation
When using `partType` parameter, openscadgen can generate part letters (A, B, C...) for multi-part designs.

### Parameter Processing
Use `ignore_param_when_processing` to exclude parameters from being passed to OpenSCAD but still use them in file naming:

```toml
[[openscadgen.input_paths]]
path = "./design.scad"
ignore_param_when_processing = "cup_holders_mode"
```

This is useful when you want to use a parameter for organization but it's not needed by the OpenSCAD file itself.

The config file format uses [toml](https://toml.io/en/) and some examples are include below and more are in the 'examples' directory

To run the tool, you need to have a config file and OpenSCAD executable file available on PATH.

To run, use: 
```
./openscadgen -c path/to/config.toml
```
(where config.toml is the openscadgen config file you want to generate for)


An [example config file](./examples/screw-mounted-clip/config.toml) is provided for the excellent "[Command Strip Parametric Broom Handle Clip](https://www.printables.com/model/516117-parametric-broom-handle-holder-openscad-command-st/related)" by [Brian Khuu](https://briankhuu.com/). 

# Installation 
Available via the [github release page](https://github.com/kiwikid/openscadgen/releases)

## Docker Usage

You can run openscadgen using Docker without installing Go or OpenSCAD locally.

### Build the Docker image

```bash
docker build -t openscadgen .
```

### Run a single config file

Mount your project directory and run the tool:

```bash
docker run --rm -v /path/to/your/project:/app/project openscadgen openscadgen -c /app/project/config.toml
```

### Run in server mode

```bash
docker run --rm -p 6767:6767 -v /path/to/your/projects:/app/examples openscadgen openscadgen -sf /app/examples -p 6767 --server-address 0.0.0.0
```

This starts the server at http://localhost:6767 serving the projects in the mounted directory.

### Notes

- The Docker image includes OpenSCAD and the BOSL2 library.
- Use `--rm` to automatically remove the container after it finishes.
- The `-v` flag mounts your local directory into the container.

# AI & MCP

OpenSCADGen can publish a [Model Context Protocol](https://modelcontextprotocol.io) server so external clients — Claude Desktop, Cursor, VS Code, MCP Inspector, or your own agent — can drive a project directly. There is no built-in chat UI and no API key: the client owns the model, OpenSCADGen owns the tools.

```bash
# HTTP transport (streamable HTTP on POST /mcp)
openscadgen -s -sf ./projects -ai -p 7203

# stdio transport (the client spawns the process)
openscadgen -ai -mcp-stdio -mcp-root ./projects
```

With `-ai` the server also exposes a read-only settings page at <http://127.0.0.1:7203/ai> showing the effective configuration, the advertised tool list, and copy-ready client snippets. The page is mounted only when `-ai` (or `OPENSCADGEN_AI_ON=true`) is set; `/ai`, `/mcp`, and the `/api/mcp/process/` progress endpoint return 404 otherwise.

## Configuration

Nothing is persisted. Every setting resolves from a flag, then an environment variable, then a default; restart the server to change it.

| Flag | Environment | Default | Purpose |
| --- | --- | --- | --- |
| `-ai` | `OPENSCADGEN_AI_ON` | off | Enable the MCP server and settings page |
| `-mcp-stdio` | — | off | Serve MCP on stdin/stdout instead of HTTP (implies `-ai`, logs move to stderr) |
| `-mcp-root` | `OPENSCADGEN_MCP_ROOT` | server folder, else working directory | Comma-separated project folders tools may access |
| `-mcp-detail` | `OPENSCADGEN_MCP_DETAIL` | `normal` | Default `response_detail`: `summary`, `normal`, or `full` |
| `-mcp-read-only` | `OPENSCADGEN_MCP_READ_ONLY` | off | Advertise read-only tools only |
| `-mcp-allow-push` | `OPENSCADGEN_MCP_ALLOW_PUSH` | off | Advertise `git_push` |

Paths passed to tools must stay inside a configured root. Escapes — including through symlinks and junctions — are refused. With several roots, a relative path must identify exactly one root or the call is rejected as ambiguous.

## Tools

| Tool | Access | Notes |
| --- | --- | --- |
| `list_projects` | read | Finds folders containing `config.toml` under the roots |
| `list_project_files` | read | Depth ≤ 4, capped listing, skips `.git`/`node_modules` |
| `read_project_file` | read | Bounded read: default 20 KB, max 200 KB, `offset`/`max_bytes` for paging |
| `edit_project_file` | write | Exact-text replacement; validates `config.toml` before writing |
| `write_project_file` | write | Create/overwrite; validates `config.toml` before writing |
| `process_config` | write | Validate synchronously, then render as a background job; returns `job_id` + `progress_url` |
| `process_status` | read | Poll a render job over any transport |
| `git_status` | read | Scoped to the root |
| `git_diff` | read | `summary` returns the stat line, `normal` the patch |
| `git_commit` | write | Stages tracked changes by default; pass `paths` to include new files |
| `git_push` | push | Only advertised with `-mcp-allow-push` |

A `config.toml` edit or write is validated **before** it touches disk: an invalid TOML edit is refused and the existing file is left untouched. `process_config` runs validation as a separate stage, so a broken config reports validation errors without invoking the renderer, and it checks that every expected STL/PNG exists and is non-empty after the run. A run where instances failed is reported with `ok: false` and an `error`, never a false success.

### Render progress

Rendering is asynchronous. `process_config` validates and starts the render, then returns immediately with a `job_id`, a `progress_url`, and `progress_tool: "process_status"`:

```json
{
  "ok": true,
  "status": "running",
  "done": false,
  "job_id": "f90d43be-2898-4492-899e-02e90a83b428",
  "config_path": "cup/config.toml",
  "started_at": "2026-11-21T10:52:03Z",
  "elapsed_ms": 213,
  "current_instance": "slow_10",
  "completed_instances": 0,
  "total_instances": 2,
  "progress_url": "http://127.0.0.1:7203/api/mcp/process/f90d43be-2898-4492-899e-02e90a83b428",
  "progress_tool": "process_status"
}
```

Poll `progress_url` with `GET` (add `?detail=summary|normal|full` to choose the result detail):

```bash
curl -s http://127.0.0.1:7203/api/mcp/process/<job_id>
```

The same snapshot is available over any transport through the `process_status` tool, which is the only option in stdio mode because no HTTP server is mounted:

```json
{ "name": "process_status", "arguments": { "job_id": "<job_id>" } }
```

The snapshot always carries `started_at`, `elapsed_ms`, `current_instance`, `completed_instances`, and `total_instances`. When `done` becomes `true`, `status` is `done`, `ok` reflects the render outcome, and `result` holds the same payload a blocking call would have returned (artifacts, issues, counts).

Clients that would rather block can pass `wait_seconds` (max 300); the call then returns the finished result, or the running snapshot if the timeout expires first:

```json
{ "path": "cup/config.toml", "wait_seconds": 120 }
```

Finished jobs are kept for 15 minutes (bounded to 32 jobs) so a slow poll after a long render still succeeds; unknown or expired ids return 404 / a tool error.

### `response_detail`

Every tool accepts an optional `response_detail` that overrides the server default for that call:

- `summary` — just the outcome: counts, paths, commit hash. Cheapest for an agent loop.
- `normal` — adds diagnostics: artifact lists, issues, diffs, command output.
- `full` — adds per-instance detail and provenance.

Failures always return full detail regardless of the requested level, so a client never has to re-run a failed call to diagnose it.

## Client setup

HTTP clients (Cursor, VS Code, MCP Inspector):

```json
{
  "mcpServers": {
    "openscadgen": {
      "url": "http://127.0.0.1:7203/mcp"
    }
  }
}
```

Clients that spawn a process (Claude Desktop):

```json
{
  "mcpServers": {
    "openscadgen": {
      "command": "C:\\path\\to\\openscadgen.exe",
      "args": ["-ai", "-mcp-stdio", "-mcp-root", "C:\\projects"]
    }
  }
}
```

For clients that only speak stdio but should reach a running server, bridge with
`npx -y mcp-remote http://127.0.0.1:7203/mcp`.

## Security

The MCP endpoint has **no authentication**. It binds to `127.0.0.1` by default; if you set `--server-address` to a non-loopback address, anyone who can reach the port can read, edit, commit, and push within the configured roots. Put an authenticating proxy in front of it, or keep it on loopback.

Because a client drives the tools, treat project content as untrusted input: a prompt-injected README or `.scad` file could ask a model to edit files or push. Use `-mcp-read-only` when you only need inspection, and leave `-mcp-allow-push` off unless you need it.

File tools refuse paths inside `.git`, so a model cannot rewrite `.git/config` or install hooks. Git tools only ever run real `git` subcommands with prompts disabled.

# Command Line Options

## Basic Usage
```bash
-c, -config string     Path to config file (.toml)
-ow, -overwrite        Overwrite existing files
-v, -version          Show version information
-h, -man              Display help message
```

## Quality & Performance
```bash
-hq                    Set high quality (fn = 200)
-lq                    Set low quality (fn = 20)
-fn int               Override the default fn value
-sr, -skip-render     Don't run a render before export
```

## Processing Control
```bash
-n int                Maximum number of instances to process
-r, -regex string     Regex pattern to filter instances by name
-np, -no-processing   Dry-run mode - check config without processing
-coe, -continue-on-error  Continue on error
```

## Output Options
```bash
-sd, -skip-docs       Skip generating a README.md file
-el, -include-export-log-file  Include export log in README.md
-pid                  Include optional_part_id_letter variable
-fi                   Set build info in file attributes (default: true)
-br, -browser-report  Embed the OpenSCAD WebAssembly browser renderer in report.html
```

## Project Management
```bash
-i, -init string      Initialize a new project
-ie, -init-extended string  Initialize with BOSL2 and renderSlicing support
-ix, -init-explainer string  Initialize with a more verbose starter config
-no-input             Skip interactive init prompts
-custom-openscad-command string  Custom OpenSCAD command to use
-dangerously-skip-permissions    Allow unsafe OpenSCAD settings such as custom command paths and extra command arguments
```

Use `-dangerously-skip-permissions` only when you trust the config and the OpenSCAD binary you are about to run. The main risks are:
- Arbitrary code execution if a custom command points at a malicious binary or wrapper script
- Privilege escalation if the binary path resolves to a more privileged interpreter or launcher
- Shell injection risk from custom command strings and command-line argument pass-through
- Unsafe file access or exfiltration from OpenSCAD scripts that can read local resources through imported files or libraries

The flag is intentionally narrow. It does not disable normal validation, and it should only be needed for features that require trusted execution.

### Project Initialization
Use `-init` to quickly create a new project structure. On a first project, the CLI will prompt to use the explainer starter template and will auto-accept after 10 seconds:
```bash
./openscadgen -init my-new-project
```

This creates a basic config file and directory structure to get you started quickly.

If you want the explanatory starter config directly, use:
```bash
./openscadgen -init-explainer my-new-project
```

## AI & MCP Options
```bash
-ai                   Enable the MCP server and AI settings page in server mode
-mcp-stdio            Serve MCP on stdin/stdout instead of HTTP (implies -ai)
-mcp-root string      Comma-separated project folders MCP tools may access
-mcp-detail string    Default response detail: summary, normal, or full (default normal)
-mcp-read-only        Advertise read-only tools only
-mcp-allow-push       Advertise the git_push tool
```

See [AI & MCP](#ai--mcp) for the tool list, client configuration, and security notes.

## Debug & Verbosity
```bash
-d, -debug            Debug mode, more output
-q, -quiet            Quiet mode, no log output
```

## Troubleshooting

Message:
```
./openscadgen -c examples/cup-holder-plus/config.toml
exec: Failed to execute process: './openscadgen' the file could not be run by the operating system.
```
Ensure you have downloaded the correct version of the tool from the releases page (i.e. darwin_arm64, linux_arm64, etc)

Message:
```
./openscadgen -c examples/cup-holder-plus/config.toml
fish: Job 1, './openscadgen -c examples/cup-h…' terminated by signal SIGKILL (Forced quit)
```
Ensure you have allowed openscadgen to run (in Privacy & Security settings)


# Config file format

The config is TOML and has one `[openscadgen]` table. Paths are resolved from
the `config.toml` file. `input_path` is the compact form for one source;
`[[openscadgen.input_paths]]` is the form to use for multiple sources or
per-source settings. A config must provide a `name` and at least one input
path; `version` defaults to `v0.1` and `run_type` defaults to
`clearAndCreate`.

```toml
[openscadgen]
name = "screw-mounted-clip"
description = "A parametric screw-mounted clip"
input_path = "./clip.scad"
export_name_format = "clip-{handle_diameter}mm-{handle_offset}mm"
version = "v1.0"

[[openscadgen.instances]]
name = "standard"
params = { handle_diameter = "5,10", handle_offset = "5,15" }

[[openscadgen.images]]
name = "nice"
```

`params` values containing commas expand into combinations. Use TOML numbers,
booleans, strings, or arrays as appropriate. Parameter names must be valid
OpenSCAD identifiers. Put every parameter that varies between instances in
`export_name_format`, otherwise the run can produce duplicate file paths.

## Configuration reference

### `[openscadgen]`

| Field | What it does |
| --- | --- |
| `name` | Required project name. It is shown in generated documentation and reports. |
| `description` | Optional project description for generated documentation and the UI. |
| `input_path` | One `.scad` input file, relative to this config. Do not combine it with `input_paths`. |
| `input_paths` | Array of input definitions. Use this instead of `input_path` for multiple files or per-file options. |
| `output_path` | Legacy accepted field. The current exporter writes beside the config/input under its versioned `export/` structure; this field does not select a different export root. |
| `version` | Version directory label, for example `v1.0`; defaults to `v0.1`. |
| `export_name_format` | Required filename/path template for STL exports. Supports `{designFileName}`, `{instanceName}`, and `{paramName}` for any applied parameter. It may include subdirectories. |
| `global_params` | Parameter map merged into every instance. Later per-input, parameter-set, and instance values can override it. |
| `param_sets` | Array of named reusable parameter maps, selected by `param_sets` on an input, instance, or targeted image. |
| `instances` | Array of instance definitions. Each definition expands into one output for every parameter combination. If omitted, one empty `default` instance is generated. |
| `images` | Array of cameras applied to matching instances to create PNG previews. |
| `image` | Array of targeted, one-off image renders. It reuses at most one matching instance, or creates an image-only report entry. |
| `tests` | Array of assertions run against each generated part. |
| `export_text_sizing` | Array of font/parameter pairs used by text-sizing export support. |
| `run_type` | `clearAndCreate` (default) clears the previous version folder after confirmation; `appendOrOverwrite` keeps it and skips that deletion. |
| `no_part_id_letter` | Legacy accepted field; current processing assigns part letters independently, so this does not suppress them. |
| `dont_use_manifold` | Legacy accepted field; the Manifold backend switch is currently not passed to OpenSCAD, so this has no effect. |
| `custom_openscad_output_format` | Changes the OpenSCAD export extension/format, for example `stl`. The command-line output-format flag overrides it. |
| `custom_openscad_args` | Extra OpenSCAD arguments. This requires `--dangerously-skip-permissions` because arguments are executed locally. |
| `export_image_quality` | Legacy accepted field; current image quality is controlled through CLI quality options rather than this config field. |

### `[[openscadgen.tests]]`

| Field | What it does |
| --- | --- |
| `type` | `is-flat` checks an STL contact face. `no-holes` checks a rendered view for enclosed background, which indicates a see-through hole. `is-different` renders two parameter variants and requires a visible difference. `min-color-pixels` requires a visible target-colour area in one rendered view. `color-in-bounding-box` requires that colour within a specified square. |
| `camera` | Required for `no-holes`. The named preset camera to render and inspect; it is generated automatically if it is not already a preview. |
| `image_size` | Optional size for an automatically generated `no-holes` camera. |
| `min_hole_pixels` | Minimum enclosed-background component size that fails `no-holes`; defaults to `1`. |
| `instance_target` | Required for `is-different`: the name of one `[[openscadgen.instances]]` entry. Both renders inherit its configuration. |
| `params` | Optional overrides for the first `is-different` render. |
| `different_params` | Optional overrides for the second `is-different` render. |
| `min_different_pixels` | Minimum differing pixels required by `is-different`; defaults to `1`. |
| `color` | Required `#RRGGBB` target colour for `min-color-pixels`. |
| `min_color_pixels` | Minimum pixels near `color` required by `min-color-pixels`; defaults to `1`. |
| `bounding_box` | Required by `color-in-bounding-box`: `[left, top, size]` in exported-image pixels. |

```toml
[[openscadgen.tests]]
type = "no-holes"
camera = "top"
```

Outer background remains valid: only a background region that is fully enclosed by a rendered object fails this test.

`is-different` is useful for a modifier or mode that must visibly change the export without repeating its base configuration:

```toml
[[openscadgen.tests]]
type = "is-different"
instance_target = "preview"
camera = "top"
params = { modifier = false }
different_params = { modifier = true }
min_different_pixels = 10
```

The two images are rendered from the target instance's parameters plus the respective overrides. The target must resolve to one image for each side of the comparison.

`min-color-pixels` is useful when a test-only scene colours a modifier and the camera must show enough of it through its matching cutout:

```toml
[[openscadgen.tests]]
type = "min-color-pixels"
instance_target = "top-colour-alignment"
camera = "side-near"
color = "#ff0000"
min_color_pixels = 10000
```

`color-in-bounding-box` confines the colour check to a square. The report shows a generated diagnostic image with a green border for a passing region and a red border for a failing one. Several checks against the same render share that one annotated image.

```toml
[[openscadgen.tests]]
type = "color-in-bounding-box"
instance_target = "top-colour-alignment"
camera = "side-near"
color = "#ff0000"
bounding_box = [120, 80, 64] # left, top, square side length in pixels
min_color_pixels = 1
```

### `[[openscadgen.input_paths]]`

| Field | What it does |
| --- | --- |
| `path` | Required path to the `.scad` source. |
| `raw_openscad_file` | Inline OpenSCAD source accepted by the schema. Current instance generation still requires `path`; use a real source file for reliable processing. |
| `raw_openscad_file_name` | Intended filename for inline source; currently only stored with the input definition. |
| `export_name_format` | Naming-template override for this source. |
| `param_sets` | Comma-separated parameter-set names to merge into this source's instances. |
| `params` | Parameters applied to every instance generated from this source. |
| `skip_images` | Do not create preview images for this source. |
| `ignore_param_when_processing` | Comma-separated parameter names omitted from OpenSCAD definitions for this source. They can still be used in filenames. |

### `[[openscadgen.param_sets]]`

| Field | What it does |
| --- | --- |
| `name` | Name selected by an input, instance, or targeted image's comma-separated `param_sets` list. |
| `params` | Parameter map contributed whenever that set is selected. |

### `[[openscadgen.instances]]`

| Field | What it does |
| --- | --- |
| `name` | Optional instance label, available as `{instanceName}` in export names and shown in reports. |
| `description` | Optional description shown for this instance in reports. |
| `export_name_format` | Naming-template override for this instance; it takes precedence over an input or top-level template. |
| `params` | Per-instance parameters. Comma-separated string values expand to combinations unless listed in `ignore_comma_in_params`. |
| `param_sets` | Comma-separated names of reusable parameter sets to apply. |
| `params_numbered` | Comma-separated string values become numbered parameter names (`slot1`, `slot2`, ...), useful for repeated OpenSCAD arguments. |
| `ignore_comma_in_params` | Parameter names whose comma-containing strings are literal text rather than lists to expand. |
| `direct_array_params` | Parameter names that should remain OpenSCAD arrays rather than be expanded into separate instances. |
| `images` | Array of cameras used only for this instance. |
| `skip_images` | Do not create preview images for this instance. |

### `[[openscadgen.images]]` and `[[openscadgen.instances.images]]`

Both camera arrays support the same fields. Top-level `images` applies to every
matching instance; instance-level `images` applies only to its parent.

| Field | What it does |
| --- | --- |
| `name` | Camera preset name, such as `top`, `bottom`, `front`, `back`, `left`, `right`, `side`, or `nice`. `nice-100` through `nice-1000` use the `nice` angle at that distance. |
| `type` | Top-level images only: omit it or use `instance` for normal per-instance PNGs; use `all_instances` to additionally create one montage for that camera. |
| `export_name_format` | Image-path template override. It supports the same placeholders as STL export names. |
| `coord` | Optional custom seven-value OpenSCAD camera string (`translation,rotation,distance`); it overrides the preset position. |
| `image_size` | Optional per-camera dimensions, for example `1200x1200`. |
| `param_filter` | Parameter map that must match an instance before this camera is used. |

### `[[openscadgen.image]]` (targeted image)

| Field | What it does |
| --- | --- |
| `name` | Required camera preset name for this one render. |
| `instance_name` | Name used in the report if the image requires an image-only instance. |
| `input_path` | Required to disambiguate the source when the design has more than one input path. |
| `params` | Parameters used to select an existing instance or create the image-only instance. They must match at most one configured instance. |
| `param_sets` | Comma-separated parameter-set names to merge into this render. |
| `export_name_format` | Optional naming-template override for an image-only instance. |
| `coord` | Optional custom seven-value OpenSCAD camera string; otherwise the `name` preset is used. |
| `image_size` | Optional dimensions for this one image. |

### `[[openscadgen.export_text_sizing]]`

| Field | What it does |
| --- | --- |
| `font` | Font family used by the text-sizing entry. |
| `param_name` | OpenSCAD parameter name to which that text-sizing entry applies. |

For a terminal-friendly reference kept alongside the implementation, run
`./openscadgen -config-options -topic images` (omit `-topic` for every option).


# Development
```bash
# build
go build .

# test
go test -v ./tests

# run
./openscadgen -c ./examples/screw-mounted-clip/config.toml

# test



# release   
# bump version in main.go
git commit -m "New and improved version"
git tag "v[NEW_VERSION_HERE]-alpha"  
git push && git push --tags
```

## TODO/Project Ideas
- [ ] Better validation/messaging around config issues
- [ ] Add openscad parameter JSON file export
- [ ] Allow for concurrent openscad runs
- [ ] Directory generation (i.e. dynamically find and generate all the instance configs in a directory)
```
// openscadgen: handle_offset: 10, 15, 25
handle_offset = 10
```
- [ ] Allow for building a whole directory of scad files (i.e. generate all the instances in a directory)
- [ ] Split export folder into 'base' and 'has_part_letter' folders + refine part letter generation
- [ ] Allow for setting of part id in the static instance config
- [ ] Add clean-up option for old versions
- [ ] (maybe) Add configurable watch mode to automatically re-run the tool when the scad file is changed
- [ ] Tidy/Improve logging and log handling
- [ ] Tools to ease handling of openscad directly
- [ ] Tools to ease handling of common openscad external libraries (i.e. BOSL2 etc)
- [ ] (Investigate/Fix) Negative number limitation (could be CLI Param handling or openscad CLI limitiation)

- [-] Better handling of params with Spaces in them (i.e. name= "My Name" -> "My-Name.stl")
- [-] Better handling of params with Spaces in them (i.e. name= "My Name" -> "My-Name.stl")
- [-] Add a HTML export file + image generation
- [-] Warn when replacing existing stl export files
- [-] Add ability to configure ranges of parameters in config file with auto-naming (i.e. i want models for each handle_diameter from 5 to 10, with )
- [-] Allow for multi-part builds in the same config file (i.e. multiple scad files, generated with the same input parameters)
- [-] Add config file generation quickstart command (to initialize a openscadgen config file from a scad file) (`openscadgen i new-project-name`)
- [-] Configure option (-ie) to automatically create a horizontal and/or vertical cross-sectional slice of the part
- [-] Allow for exporting a set of images



## Development

For development with live reloading:

```bash
# Install air for live reloading
go install 

# Live reload both templ and go files
just air

# Live reload only templ files
just air-templ

# Live reload only go files (after templ is generated)
just air-go

# Clean and start fresh
just dev-clean
```


If you have any ideas/bugs/etc, please let me know and i'll try and fix them where possible. I do want to keep the goals of the project simple and specific
